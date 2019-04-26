// Copyright 2018 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from p2p/discover/table.go (2018/06/04).
// Modified and improved for the klaytn development.

package discover

import (
	"errors"
	"fmt"
	mrand "math/rand"
	"net"
	"sync"
	"time"

	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/networks/p2p/netutil"
)

type Simple struct {
	mutex      sync.Mutex  // protects buckets, bucket content, nursery, rand
	bucket     *bucket     // contains bonded nodes
	bucketSize int         // the max number of bucket entries
	nursery    []*Node     // bootstrap nodes
	rand       *mrand.Rand // source of randomness, periodically reseeded
	ips        netutil.DistinctNetSet

	db         *nodeDB // database of known nodes
	refreshReq chan chan struct{}
	initDone   chan struct{}
	closeReq   chan struct{}
	closed     chan struct{}

	bondmu    sync.Mutex
	bonding   map[NodeID]*bondproc
	bondslots chan struct{} // limits total number of active bonding processes

	nodeAddedHook func(*Node) // for testing

	net  transport
	self *Node // metadata of the local node
}

func newSimple(cfg *Config) (Discovery, error) {
	// If no node database was given, use an in-memory one
	db, err := newNodeDB(cfg.NodeDBPath, Version, cfg.Id)
	if err != nil {
		return nil, err
	}
	discovery := &Simple{
		net:        cfg.udp,
		db:         db,
		self:       NewNode(cfg.Id, cfg.Addr.IP, uint16(cfg.Addr.Port), uint16(cfg.Addr.Port)),
		bonding:    make(map[NodeID]*bondproc),
		bondslots:  make(chan struct{}, maxBondingPingPongs),
		refreshReq: make(chan chan struct{}),
		initDone:   make(chan struct{}),
		closeReq:   make(chan struct{}),
		closed:     make(chan struct{}),
		rand:       mrand.New(mrand.NewSource(0)),
		ips:        netutil.DistinctNetSet{Subnet: tableSubnet, Limit: tableIPLimit},
		bucketSize: int(cfg.MaxNeighborsNode),
	}

	if err := discovery.setFallbackNodes(cfg.Bootnodes); err != nil {
		return nil, err
	}
	for i := 0; i < cap(discovery.bondslots); i++ {
		discovery.bondslots <- struct{}{}
	}
	discovery.bucket = &bucket{
		ips: netutil.DistinctNetSet{Subnet: bucketSubnet, Limit: bucketIPLimit},
	}

	// Start the background expiration goroutine after loading seeds so that the search for
	// seed nodes also considers older nodes that would otherwise be removed by the
	// expiration.
	discovery.db.ensureExpirer()
	go discovery.loop()
	return discovery, nil
}

func (s *Simple) Name() string { return "SimpleDiscovery" }

// Self returns the local node.
// The returned node should not be modified by the caller.
func (s *Simple) Self() *Node {
	return s.self
}

func (s *Simple) CreateUpdateNode(n *Node) error {
	return s.db.updateNode(n)
}

func (s *Simple) GetNode(id NodeID) (*Node, error) {
	node := s.db.node(id)
	if node == nil {
		return nil, errors.New("failed to retrieve the node with the given id")
	}
	return node, nil
}

func (s *Simple) DeleteNode(id NodeID) error {
	return s.db.deleteNode(id)
}

// ReadRandomNodes is not needed on the simple discovery. So, it just returns 0.
func (s *Simple) ReadRandomNodes(buf []*Node) (n int) {
	return 0
}

// Close terminates the network listener and flushes the node database.
func (s *Simple) Close() {
	select {
	case <-s.closed:
		// already closed.
	case s.closeReq <- struct{}{}:
		<-s.closed // wait for refreshLoop to end.
	}
}

// setFallbackNodes sets the initial points of contact. These nodes
// are used to connect to the network if the table is empty and there
// are no known nodes in the database.
func (s *Simple) setFallbackNodes(nodes []*Node) error {
	for _, n := range nodes {
		if err := n.validateComplete(); err != nil {
			return fmt.Errorf("bad bootstrap/fallback node %q (%v)", n, err)
		}
	}
	s.nursery = make([]*Node, 0, len(nodes))
	for _, n := range nodes {
		cpy := *n
		// Recompute cpy.sha because the node might not have been
		// created by NewNode or ParseNode.
		cpy.sha = crypto.Keccak256Hash(n.ID[:])
		s.nursery = append(s.nursery, &cpy)
	}
	return nil
}

// isInitDone returns whether the table's initial seeding procedure has completed.
func (s *Simple) isInitDone() bool {
	select {
	case <-s.initDone:
		return true
	default:
		return false
	}
}

// Resolve searches for a specific node with the given ID.
// It returns nil if the node could not be found.
func (s *Simple) Resolve(targetID NodeID) *Node {
	for _, val := range s.bucket.entries {
		if val.ID == targetID {
			return val
		}
	}

	for _, val := range s.bucket.replacements {
		if val.ID == targetID {
			return val
		}
	}
	return nil
}

func (s *Simple) LookupByType(targetID NodeID, t DiscoveryType) []*Node {
	//TODO-Klaytn-Node implement this method
	return nil
}

func (s *Simple) Lookup(targetID NodeID) []*Node {
	return s.lookup(targetID, true)
}

func (s *Simple) lookup(targetID NodeID, refreshIfEmpty bool) []*Node {
	if len(s.nursery) == 0 {
		return nil
	}

	var (
		result []*Node
		reply  = make(chan []*Node)
	)

	for _, rn := range s.nursery {
		n := rn
		go func() {
			// Find potential neighbors to bond with
			r, err := s.net.findnode(n.ID, n.addr(), s.self.ID)
			if err != nil {
				// Bump the failure counter to detect and evacuate non-bonded entries
				fails := s.db.findFails(n.ID) + 1
				s.db.updateFindFails(n.ID, fails)
				logger.Trace("Bumping findnode failure counter", "id", n.ID, "failcount", fails)

				if fails >= maxFindnodeFailures {
					logger.Trace("Too many findnode failures, dropping", "id", n.ID, "failcount", fails)
					//s.delete(n)
					//TODO-Klaytn-Bootnode it always return RPC time out error. The delete operation is commented out for now, but need to find out why.
				}
			}
			reply <- s.bondall(r)
		}()
	}

	for range s.nursery {
		for _, n := range <-reply {
			if len(result) < s.bucketSize {
				result = append(result, n)
			}
		}
	}

	return result
}

func (s *Simple) refresh() <-chan struct{} {
	done := make(chan struct{})
	select {
	case s.refreshReq <- done:
	case <-s.closed:
		close(done)
	}
	return done
}

// loop schedules refresh, revalidate runs and coordinates shutdown.
func (s *Simple) loop() {
	var (
		revalidate     = time.NewTimer(s.nextRevalidateTime())
		copyNodes      = time.NewTicker(copyNodesInterval)
		revalidateDone = make(chan struct{})
		refreshDone    = make(chan struct{})         // where doRefresh reports completion
		waiting        = []chan struct{}{s.initDone} // holds waiting callers while doRefresh runs
	)
	defer revalidate.Stop()
	defer copyNodes.Stop()

	// Start initial refresh.
	go s.doRefresh(refreshDone)

loop:
	for {
		select {
		case req := <-s.refreshReq:
			waiting = append(waiting, req)
			if refreshDone == nil {
				refreshDone = make(chan struct{})
				go s.doRefresh(refreshDone)
			}
		case <-refreshDone:
			for _, ch := range waiting {
				close(ch)
			}
			waiting, refreshDone = nil, nil
		case <-revalidate.C:
			go s.doRevalidate(revalidateDone)
		case <-revalidateDone:
			revalidate.Reset(s.nextRevalidateTime())
		case <-copyNodes.C:
			go s.persistBondedNodes()
		case <-s.closeReq:
			break loop
		}
	}

	if s.net != nil {
		s.net.close()
	}
	if refreshDone != nil {
		<-refreshDone
	}
	for _, ch := range waiting {
		close(ch)
	}
	s.db.close()
	close(s.closed)
}

func (s *Simple) GetBucketEntries() []*Node {
	return s.bucket.entries
}

func (s *Simple) GetReplacements() []*Node {
	return s.bucket.replacements
}

// doRefresh performs a lookup to bootnodes in order to update the bucket.
func (s *Simple) doRefresh(done chan struct{}) {
	defer close(done)

	// Run self lookup to discover new neighbor nodes.
	s.lookup(s.self.ID, false)
}

// doRevalidate checks that the last node in a random bucket is still live
// and replaces or deletes the node if it isn't.
func (s *Simple) doRevalidate(done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	for _, n := range s.nursery {
		if err := s.ping(n.ID, n.addr()); err != nil {
			logger.Debug("[Discovery] Sending ping to bootnode returned error.")
		}
	}

	last := s.nodeToRevalidate()
	if last == nil {
		// No non-empty bucket found.
		return
	}

	// Ping the selected node and wait for a pong.
	err := s.ping(last.ID, last.addr())

	s.mutex.Lock()
	defer s.mutex.Unlock()
	b := s.bucket
	if err == nil {
		// The node responded, move it to the front.
		b.bump(last)
		return
	}
	// No reply received, pick a replacement or delete the node if there aren't
	// any replacements.
	if r := s.replace(b, last); r != nil {
		logger.Debug("[Discovery] Replaced dead node", "id", last.ID, "ip", last.IP, "r", r.ID, "rip", r.IP)
	} else {
		logger.Debug("[Discovery] Removed dead node", "id", last.ID, "ip", last.IP)
	}
	s.lookup(s.self.ID, false)
}

// nodeToRevalidate returns the last node in a random, non-empty bucket.
func (s *Simple) nodeToRevalidate() *Node {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	b := s.bucket
	if len(b.entries) > 0 {
		last := b.entries[len(b.entries)-1]
		return last
	}
	return nil
}

func (s *Simple) nextRevalidateTime() time.Duration {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return time.Duration(s.rand.Int63n(int64(revalidateInterval)))
}

// persistBondedNodes adds nodes from the table to the database if they have been in the table
// longer then minTableTime.
func (s *Simple) persistBondedNodes() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, n := range s.bucket.entries {
		// TODO-Klaytn-bootnode the original code has the interval for updating node info into DB when it is greater than seenMinTableTime.
		s.db.updateNode(n)
	}
}

// randomNodes returns the n nodes that are randomly picked in the table.
func (s *Simple) randomNodes(nresults int) []*Node {
	var result []*Node
	indices := s.rand.Perm(len(s.bucket.entries))
	for guard, idx := range indices {
		if guard == nresults {
			break
		}
		result = append(result, s.bucket.entries[idx])
	}
	return result
}

func (s *Simple) RetrieveNodes(target common.Hash, nresults int) []*Node {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.randomNodes(s.bucketSize)
}

func (s *Simple) len() (n int) {
	return len(s.bucket.entries)
}

// bondall bonds with all given nodes concurrently and returns
// those nodes for which bonding has probably succeeded.
func (s *Simple) bondall(nodes []*Node) (result []*Node) {
	rc := make(chan *Node, len(nodes))
	for i := range nodes {
		go func(n *Node) {
			nn, _ := s.bond(false, n.ID, n.addr(), n.TCP)
			rc <- nn
		}(nodes[i])
	}
	for range nodes {
		if n := <-rc; n != nil {
			result = append(result, n)
		}
	}
	return result
}

// bond ensures the local node has a bond with the given remote node.
// It also attempts to insert the node into the table if bonding succeeds.
// The caller must not hold tab.mutex.
//
// A bond is must be established before sending findnode requests.
// Both sides must have completed a ping/pong exchange for a bond to
// exist. The total number of active bonding processes is limited in
// order to restrain network use.
//
// bond is meant to operate idempotently in that bonding with a remote
// node which still remembers a previously established bond will work.
// The remote node will simply not send a ping back, causing waitping
// to time out.
//
// If pinged is true, the remote node has just pinged us and one half
// of the process can be skipped.
func (s *Simple) bond(pinged bool, id NodeID, addr *net.UDPAddr, tcpPort uint16) (*Node, error) {
	if id == s.self.ID {
		return nil, errors.New("is self")
	}
	if pinged && !s.isInitDone() {
		return nil, errors.New("still initializing")
	}
	// Start bonding if we haven't seen this node for a while or if it failed findnode too often.
	node, fails := s.db.node(id), s.db.findFails(id)
	age := time.Since(s.db.bondTime(id))
	var result error
	if fails > 0 || age > nodeDBNodeExpiration {
		logger.Trace("Starting bonding ping/pong", "id", id, "known", node != nil, "failcount", fails, "age", age)

		s.bondmu.Lock()
		w := s.bonding[id]
		if w != nil {
			// Wait for an existing bonding process to complete.
			s.bondmu.Unlock()
			<-w.done
		} else {
			// Register a new bonding process.
			w = &bondproc{done: make(chan struct{})}
			s.bonding[id] = w
			s.bondmu.Unlock()
			// Do the ping/pong. The result goes into w.
			s.pingpong(w, pinged, id, addr, tcpPort)
			// Unregister the process after it's done.
			s.bondmu.Lock()
			delete(s.bonding, id)
			s.bondmu.Unlock()
		}
		// Retrieve the bonding results
		result = w.err
		if result == nil {
			node = w.n
		}
	}
	// Add the node to the table even if the bonding ping/pong
	// fails. It will be relaced quickly if it continues to be
	// unresponsive.
	if node != nil {
		s.add(node)
		s.db.updateFindFails(id, 0)
		lenEntries := len(s.GetBucketEntries())
		lenReplacements := len(s.GetReplacements())
		bucketEntriesGauge.Update(int64(lenEntries))
		bucketReplacementsGauge.Update(int64(lenReplacements))
	}
	return node, result
}

func (s *Simple) pingpong(w *bondproc, pinged bool, id NodeID, addr *net.UDPAddr, tcpPort uint16) {
	// Request a bonding slot to limit network usage
	<-s.bondslots
	defer func() { s.bondslots <- struct{}{} }()

	// Ping the remote side and wait for a pong.
	if w.err = s.ping(id, addr); w.err != nil {
		close(w.done)
		return
	}
	if !pinged {
		// Give the remote node a chance to ping us before we start
		// sending findnode requests. If they still remember us,
		// waitping will simply time out.
		s.net.waitping(id)
	}
	// Bonding succeeded, update the node database.
	w.n = NewNode(id, addr.IP, uint16(addr.Port), tcpPort)
	close(w.done)
}

// ping a remote endpoint and wait for a reply, also updating the node
// database accordingly.
func (s *Simple) ping(id NodeID, addr *net.UDPAddr) error {
	s.db.updateLastPing(id, time.Now())
	if err := s.net.ping(id, addr); err != nil {
		return err
	}
	s.db.updateBondTime(id, time.Now())
	return nil
}

// add attempts to add the given node its corresponding bucket. If the
// bucket has space available, adding the node succeeds immediately.
// Otherwise, the node is added if the least recently active node in
// the bucket does not respond to a ping packet.
//
// The caller must not hold tab.mutex.
func (s *Simple) add(new *Node) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	b := s.bucket
	if !s.bumpOrAdd(b, new) {
		// Node is not in table. Add it to the replacement list.
		s.addReplacement(b, new)
	}
}

// stuff adds nodes the table to the end of their corresponding bucket
// if the bucket is not full. The caller must not hold tab.mutex.
func (s *Simple) stuff(nodes []*Node) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, n := range nodes {
		if n.ID == s.self.ID {
			continue // don't add self
		}
		b := s.bucket
		if len(b.entries) < s.bucketSize {
			s.bumpOrAdd(b, n)
		}
	}
}

// delete removes an entry from the node table (used to evacuate
// failed/non-bonded discovery peers).
func (s *Simple) delete(node *Node) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.deleteInBucket(s.bucket, node)
}

func (s *Simple) addIP(b *bucket, ip net.IP) bool {
	if netutil.IsLAN(ip) {
		return true
	}
	if !s.ips.Add(ip) {
		logger.Debug("IP exceeds table limit", "ip", ip)
		return false
	}
	if !b.ips.Add(ip) {
		logger.Debug("IP exceeds bucket limit", "ip", ip)
		s.ips.Remove(ip)
		return false
	}
	return true
}

func (s *Simple) removeIP(b *bucket, ip net.IP) {
	if netutil.IsLAN(ip) {
		return
	}
	s.ips.Remove(ip)
	b.ips.Remove(ip)
}

func (s *Simple) addReplacement(b *bucket, n *Node) {
	for _, e := range b.replacements {
		if e.ID == n.ID {
			return // already in list
		}
	}
	if !s.addIP(b, n.IP) {
		return
	}
	var removed *Node
	b.replacements, removed = pushNode(b.replacements, n, maxReplacements)
	if removed != nil {
		s.removeIP(b, removed.IP)
	}
}

// replace removes n from the replacement list and replaces 'last' with it if it is the
// last entry in the bucket. If 'last' isn't the last entry, it has either been replaced
// with someone else or became active.
func (s *Simple) replace(b *bucket, last *Node) *Node {
	if len(b.entries) == 0 || b.entries[len(b.entries)-1].ID != last.ID {
		// Entry has moved, don't replace it.
		return nil
	}
	// Still the last entry.
	if len(b.replacements) == 0 {
		s.deleteInBucket(b, last)
		return nil
	}
	r := b.replacements[s.rand.Intn(len(b.replacements))]
	b.replacements = deleteNode(b.replacements, r)
	b.entries[len(b.entries)-1] = r
	s.removeIP(b, last.IP)
	return r
}

// bumpOrAdd moves n to the front of the bucket entry list or adds it if the list isn't
// full. The return value is true if n is in the bucket.
func (s *Simple) bumpOrAdd(b *bucket, n *Node) bool {
	if b.bump(n) {
		return true
	}
	if len(b.entries) >= s.bucketSize || !s.addIP(b, n.IP) {
		return false
	}
	b.entries, _ = pushNode(b.entries, n, s.bucketSize)
	b.replacements = deleteNode(b.replacements, n)
	n.addedAt = time.Now()
	if s.nodeAddedHook != nil {
		s.nodeAddedHook(n)
	}
	return true
}

func (s *Simple) deleteInBucket(b *bucket, n *Node) {
	b.entries = deleteNode(b.entries, n)
	s.removeIP(b, n.IP)
}

func (s *Simple) hasBond(id NodeID) bool {
	return s.db.hasBond(id)
}

func (s *Simple) randNode(nodes []*Node) *Node {
	rIdx := s.rand.Intn(len(nodes))
	if rIdx < 0 {
		return nil
	}
	return nodes[rIdx]
}
