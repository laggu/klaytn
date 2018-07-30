package gxdb

import (
	"flag"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

var mmap = flag.Bool("vlog_mmap", true, "Specify if value log must be memory-mapped")


func genTempDirForTestDB(b *testing.B) string {
	dir, err := ioutil.TempDir("", "klaytn-db-bench")
	if err != nil {
		b.Fatalf("cannot create temporary directory: %v", err)
	}
	return dir
}

func getKlayLDBOptions() *opt.Options {
	return getLDBOptions(128, 128)
}

func getKlayLDBOptionsX(x int) *opt.Options {
	opts := getKlayLDBOptions()
	opts.WriteBuffer *= x
	opts.BlockCacheCapacity *= x
	opts.OpenFilesCacheCapacity *= x

	return opts
}


func getKlayLDBOptionsForGetX(x int) *opt.Options {
	opts := getKlayLDBOptions()
	opts.WriteBuffer *= x
	opts.BlockCacheCapacity *= x
	opts.OpenFilesCacheCapacity *= x
	opts.DisableBlockCache = true

	return opts
}

func getKlayLDBOptionsForPutX(x int) *opt.Options {
	opts := getKlayLDBOptions()
	opts.BlockCacheCapacity *= x
	opts.BlockRestartInterval *= x

	opts.BlockSize *= x
	opts.CompactionExpandLimitFactor *= x
	opts.CompactionL0Trigger *= x
	opts.CompactionTableSize *= x

	opts.CompactionSourceLimitFactor *= x
	opts.DisableBufferPool = true
	opts.Compression = opt.DefaultCompression

	return opts
}


// readTypeFunc decides index
func benchmarkRead(b *testing.B, opts *opt.Options, valueLength, numInsertions int, readTypeFunc func(int, int) int) {
	b.StopTimer()
	b.ReportAllocs()

	dir := genTempDirForTestDB(b)
	defer os.RemoveAll(dir)

	db, err := NewLDBDatabaseWithOptions(dir, opts)
	require.NoError(b, err)
	defer db.Close()

	for i := 0; i < numInsertions ; i++ {
		bs := []byte(strconv.Itoa(i))
		db.Put(bs, randStrBytes(valueLength))
	}

	b.StartTimer()
	for i := 0; i < b.N ; i++ {
		bs := []byte(strconv.Itoa(readTypeFunc(i, numInsertions)))
		db.Get(bs)
	}
}

func randomRead(currIndex, numInsertions int) int {
	return rand.Intn(numInsertions)
}

func sequentialRead(currIndex, numInsertions int) int {
	return currIndex
}


var r = rand.New(rand.NewSource(time.Now().UnixNano()))
func zipfRead(currIndex, numInsertions int) int {
	zipf := rand.NewZipf(r, 3.14, 2.72, uint64(numInsertions))
	zipfNum := zipf.Uint64()
	return numInsertions - int(zipfNum) - 1
}

func benchmarkZipfGet(b *testing.B, opts *opt.Options, valueLength, numInsertions int) {
	benchmarkRead(b, opts, valueLength, numInsertions, zipfRead)
}

func benchmarkRandomGet(b *testing.B, opts *opt.Options, valueLength, numInsertions int) {
	benchmarkRead(b, opts, valueLength, numInsertions, randomRead)
}

func benchmarkSequentialGet(b *testing.B, opts *opt.Options, valueLength, numInsertions int) {
	benchmarkRead(b, opts, valueLength, numInsertions, sequentialRead)
}


// Length 100 & 100k Rows
func Benchmark_SequentialGet_Len100_100kRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptions()

	benchmarkSequentialGet(b, opts, 100,1000 * 100)
}

func Benchmark_SequentialGet_Len100_100kRows_KlayOptions_X2(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(2)

	benchmarkSequentialGet(b, opts, 100,1000 * 100)
}

func Benchmark_SequentialGet_Len100_100kRows_KlayOptions_X4(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(4)

	benchmarkSequentialGet(b, opts, 100,1000 * 100)
}

func Benchmark_SequentialGet_Len100_100kRows_KlayOptions_X8(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(8)

	benchmarkSequentialGet(b, opts, 100,1000 * 100)
}


// Length 100 & 1M Rows
func Benchmark_SequentialGet_Len100_1MRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptions()

	benchmarkSequentialGet(b, opts, 100,1000 * 1000)
}

func Benchmark_SequentialGet_Len100_1MRows_KlayOptions_X2(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(2)

	benchmarkSequentialGet(b, opts, 100,1000 * 1000)
}

func Benchmark_SequentialGet_Len100_1MRows_KlayOptions_X4(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(4)

	benchmarkSequentialGet(b, opts, 100,1000 * 1000)
}

func Benchmark_SequentialGet_Len100_1MRows_KlayOptions_X8(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(8)

	benchmarkSequentialGet(b, opts, 100,1000 * 1000)
}


////////////////////////////
// Length 250 & 100k Rows //
////////////////////////////
func Benchmark_SequentialGet_Len250_100kRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptions()

	benchmarkSequentialGet(b, opts, 250,1000 * 100)
}


func Benchmark_SequentialGet_Len250_100kRows_KlayOptions_X2(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(2)

	benchmarkSequentialGet(b, opts, 250,1000 * 100)
}

func Benchmark_SequentialGet_Len250_100kRows_KlayOptions_X4(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(4)

	benchmarkSequentialGet(b, opts, 250,1000 * 100)
}

func Benchmark_SequentialGet_Len250_100kRows_KlayOptions_X8(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(8)

	benchmarkSequentialGet(b, opts, 250,1000 * 100)
}


//////////////////////////
// Length 250 & 1M Rows //
//////////////////////////
func Benchmark_SequentialGet_Len250_1MRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptions()

	benchmarkSequentialGet(b, opts, 250,1000 * 1000)
}

func Benchmark_SequentialGet_Len250_1MRows_KlayOptions_X2(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(2)

	benchmarkSequentialGet(b, opts, 250,1000 * 1000)
}

func Benchmark_SequentialGet_Len250_1MRows_KlayOptions_X4(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(4)

	benchmarkSequentialGet(b, opts, 250,1000 * 1000)
}

func Benchmark_SequentialGet_Len250_1MRows_KlayOptions_X8(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(8)

	benchmarkSequentialGet(b, opts, 250,1000 * 1000)
}


//////////////////////////
// Length 500 & 1k Rows //
//////////////////////////
func Benchmark_SequentialGet_Len500_100kRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptions()

	benchmarkSequentialGet(b, opts, 500,1000 * 100)
}


///////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////// Random Get Tests Beginning ////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////

////////////////////// Length 100 & 100k Rows ///////////////////////
func Benchmark_RandomGet_Len100_100kRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptions()

	benchmarkRandomGet(b, opts, 100,1000 * 100)
}

func Benchmark_RandomGet_Len100_100kRows_KlayOptions_X2(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(2)

	benchmarkRandomGet(b, opts, 100,1000 * 100)
}

func Benchmark_RandomGet_Len100_100kRows_KlayOptions_X4(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(4)

	benchmarkRandomGet(b, opts, 100,1000 * 100)
}

func Benchmark_RandomGet_Len100_100kRows_KlayOptions_X8(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(8)

	benchmarkRandomGet(b, opts, 100,1000 * 100)
}


//////////////////////// Length 100 & 1M Rows /////////////////////////
func Benchmark_RandomGet_Len100_1MRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptions()

	benchmarkRandomGet(b, opts, 100,1000 * 1000)
}

func Benchmark_RandomGet_Len100_1MRows_KlayOptions_X2(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(2)

	benchmarkRandomGet(b, opts, 100,1000 * 1000)
}

func Benchmark_RandomGet_Len100_1MRows_KlayOptions_X4(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(4)

	benchmarkRandomGet(b, opts, 100,1000 * 1000)
}

func Benchmark_RandomGet_Len100_1MRows_KlayOptions_X8(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(8)

	benchmarkRandomGet(b, opts, 100,1000 * 1000)
}


////////////////////// Length 250 & 100k Rows ///////////////////////
func Benchmark_RandomGet_Len250_100kRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptions()

	benchmarkRandomGet(b, opts, 250,1000 * 100)
}


func Benchmark_RandomGet_Len250_100kRows_KlayOptions_X2(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(2)

	benchmarkRandomGet(b, opts, 250,1000 * 100)
}

func Benchmark_RandomGet_Len250_100kRows_KlayOptions_X4(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(4)

	benchmarkRandomGet(b, opts, 250,1000 * 100)
}

func Benchmark_RandomGet_Len250_100kRows_KlayOptions_X8(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(8)

	benchmarkRandomGet(b, opts, 250,1000 * 100)
}


//////////////////////// Length 250 & 1M Rows /////////////////////////
func Benchmark_RandomGet_Len250_1MRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptions()

	benchmarkRandomGet(b, opts, 250,1000 * 1000)
}

func Benchmark_RandomGet_Len250_1MRows_KlayOptions_X2(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(2)

	benchmarkRandomGet(b, opts, 250,1000 * 1000)
}

func Benchmark_RandomGet_Len250_1MRows_KlayOptions_X4(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(4)

	benchmarkRandomGet(b, opts, 250,1000 * 1000)
}

func Benchmark_RandomGet_Len250_1MRows_KlayOptions_X8(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(8)

	benchmarkRandomGet(b, opts, 250,1000 * 1000)
}




////////////////////// Length 500 & 100k Rows ///////////////////////
func Benchmark_RandomGet_Len500_100kRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptions()

	benchmarkRandomGet(b, opts, 500,1000 * 100)
}

func Benchmark_RandomGet_Len500_100kRows_KlayOptions_X2(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(2)

	benchmarkRandomGet(b, opts, 500,1000 * 100)
}

func Benchmark_RandomGet_Len500_100kRows_KlayOptions_X4(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(4)

	benchmarkRandomGet(b, opts, 500,1000 * 100)
}

func Benchmark_RandomGet_Len500_100kRows_KlayOptions_X8(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForGetX(8)

	benchmarkRandomGet(b, opts, 500,1000 * 100)
}



//////////////////////// Length 500 & 1M Rows /////////////////////////
func Benchmark_RandomGet_Len500_1MRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptions()

	benchmarkRandomGet(b, opts, 500,1000 * 1000)
}

func Benchmark_RandomGet_Len500_1MRows_KlayOptions_X2(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(2)

	benchmarkRandomGet(b, opts, 500,1000 * 1000)
}

func Benchmark_RandomGet_Len500_1MRows_KlayOptions_X4(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(4)

	benchmarkRandomGet(b, opts, 500,1000 * 1000)
}

func Benchmark_RandomGet_Len500_1MRows_KlayOptions_X8(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(8)

	benchmarkRandomGet(b, opts, 500,1000 * 1000)
}

///////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////// Zipf Get Tests Beginning /////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////

func Benchmark_ZipfGet_Len500_1MRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(1)

	benchmarkZipfGet(b, opts, 500,1000 * 1000)
}

func Benchmark_ZipfGet_Len500_1MRows_KlayOptions_X8(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(8)

	benchmarkZipfGet(b, opts, 500,1000 * 1000)
}

func Benchmark_ZipfGet_Len500_1MRows_KlayOptions_X10(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(10)

	benchmarkZipfGet(b, opts, 500,1000 * 1000)
}

func Benchmark_ZipfGet_Len500_1MRows_KlayOptions_X12(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(12)

	benchmarkZipfGet(b, opts, 500,1000 * 1000)
}

func Benchmark_ZipfGet_Len500_1MRows_KlayOptions_X14(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(14)

	benchmarkZipfGet(b, opts, 500,1000 * 1000)
}

func Benchmark_ZipfGet_Len500_1MRows_KlayOptions_X16(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsX(16)

	benchmarkZipfGet(b, opts, 500,1000 * 1000)
}

///////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////// Put Insertion Tests Beginning //////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////

// key = 32 bytes, value = valueLength bytes (100, 250, 500)
func benchmarkPut(b *testing.B, opts *opt.Options, valueLength, numInsertions int) {
	b.StopTimer()
	dir := genTempDirForTestDB(b)
	defer os.RemoveAll(dir)

	db, err := NewLDBDatabaseWithOptions(dir, opts)
	require.NoError(b, err)
	defer db.Close()

	b.StartTimer()
	for i := 0; i < b.N ; i++ {
		for k := 0; k < numInsertions; k++ {
			db.Put(randStrBytes(32), randStrBytes(valueLength))
		}
	}
}

func Benchmark_Put_Len500_10kRows_KlayOptions(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForPutX(1)
	benchmarkPut(b, opts, 500, 10000)
}

func Benchmark_Put_Len500_10kRows_KlayOptions_2X(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForPutX(2)
	benchmarkPut(b, opts, 500, 10000)
}

func Benchmark_Put_Len500_10kRows_KlayOptions_4X(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForPutX(4)

	benchmarkPut(b, opts, 500, 10000)
}

func Benchmark_Put_Len500_10kRows_KlayOptions_8X(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForPutX(8)

	benchmarkPut(b, opts, 500, 10000)
}

func Benchmark_Put_Len500_10kRows_KlayOptions_16X(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForPutX(16)

	benchmarkPut(b, opts, 500, 10000)
}

func Benchmark_Put_Len500_10kRows_KlayOptions_32X(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForPutX(32)

	benchmarkPut(b, opts, 500, 10000)
}

func Benchmark_Put_Len500_10kRows_KlayOptions_64X(b *testing.B) {
	b.StopTimer()
	opts := getKlayLDBOptionsForPutX(64)

	benchmarkPut(b, opts, 500, 10000)
}

///////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////// Put Insertion Tests End /////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////


///////////////////////////////////////////////////////////////////////////////////////////
//////////////////////// PARTITIONED PUT INSERTION TESTS BEGINNING ////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////

func genDirs(b *testing.B, numPartitions int) []string {
	dirs := make([]string, numPartitions, numPartitions)
	for i :=0; i < numPartitions; i++ {
		dirs[i] = genTempDirForTestDB(b)
	}
	return dirs
}

func removeDirs(dirs []string) {
	for _, dir := range dirs {
		os.RemoveAll(dir)
	}
}

func genDatabases(b *testing.B, dirs []string, opts *opt.Options) []*LDBDatabase {
	databases := make([]*LDBDatabase, len(dirs), len(dirs))
	for i := 0; i < len(dirs); i++ {
		databases[i],_ = NewLDBDatabaseWithOptions(dirs[i], opts)
	}
	return databases
}

func closeDBs(databases []*LDBDatabase) {
	for _, db := range databases {
		db.Close()
	}
}

func genKeysAndValues(valueLength, numInsertions int) ([][]byte, [][]byte) {
	keys := make([][]byte, numInsertions, numInsertions)
	values := make([][]byte, numInsertions, numInsertions)
	for i:=0; i < numInsertions; i++ {
		keys[i] = randStrBytes(32)
		values[i] = randStrBytes(valueLength)
	}
	return keys, values
}

func benchmarkPartitionedPutGoRoutine(b *testing.B, opts *opt.Options, numPartitions, valueLength, numInsertions int) {
	b.StopTimer()
	dirs := genDirs(b, numPartitions)
	defer removeDirs(dirs)

	databases := genDatabases(b, dirs, opts)
	defer closeDBs(databases)

	for i:=0; i<b.N; i++ {
		b.StopTimer()
		keys, values := genKeysAndValues(valueLength, numInsertions)
		b.StartTimer()

		var wait sync.WaitGroup
		wait.Add(numInsertions)

		for k := 0; k < numInsertions; k++ {
			go func(idx int) {
				defer wait.Done()
				if numPartitions == 1 {
					databases[0].Put(keys[idx], values[idx])
				} else {
					partition := getPartitionForTest(keys, idx, numPartitions)
					databases[partition].Put(keys[idx], values[idx])
				}
			}(k)
		}
		wait.Wait()
	}
}

func benchmarkPartitionedPutNoGoRoutine(b *testing.B, opts *opt.Options, numPartitions, valueLength, numInsertions int) {
	b.StopTimer()
	dirs := genDirs(b, numPartitions)
	defer removeDirs(dirs)

	databases := genDatabases(b, dirs, opts)
	defer closeDBs(databases)


	for i:=0; i<b.N; i++ {
		b.StopTimer()
		keys, values := genKeysAndValues(valueLength, numInsertions)
		b.StartTimer()

		for k := 0; k < numInsertions; k++ {
			if numPartitions == 1 {
				databases[0].Put(keys[k], values[k])
			} else {
				partition := getPartitionForTest(keys, k, numPartitions)
				databases[partition].Put(keys[k], values[k])
			}
		}
	}
}

// To check issue described in issue#130
//func Benchmark_PartitionedPut_Len250_100kRows_Partition4_GoRoutine(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkPartitionedPutGoRoutine(b, opts, 4, 250, 1000 * 1)
//}
//
//func Benchmark_PartitionedPut_Len250_100kRows_Partition5_GoRoutine(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkPartitionedPutGoRoutine(b, opts, 5, 250, 1000 * 1)
//}
//
//func Benchmark_PartitionedPut_Len250_100kRows_Partition6_GoRoutine(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkPartitionedPutGoRoutine(b, opts, 6, 250, 1000 * 1)
//}
//
//func Benchmark_PartitionedPut_Len250_100kRows_Partition7_GoRoutine(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkPartitionedPutGoRoutine(b, opts, 7, 250, 1000 * 1)
//}
//
//func Benchmark_PartitionedPut_Len250_100kRows_Partition8_GoRoutine(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkPartitionedPutGoRoutine(b, opts, 8, 250, 1000 * 1)
//}

// Partitioned Put with 100k Rows with length 250 bytes, GoRoutine used
func Benchmark_PartitionedPut_Len250_100kRows_Partition1_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedPutGoRoutine(b, opts, 1, 250, 1000 * 100)
}

func Benchmark_PartitionedPut_Len250_100kRows_Partition2_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedPutGoRoutine(b, opts, 2, 250, 1000 * 100)
}

func Benchmark_PartitionedPut_Len250_100kRows_Partition4_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedPutGoRoutine(b, opts, 4, 250, 1000 * 100)
}

func Benchmark_PartitionedPut_Len250_100kRows_Partition8_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedPutGoRoutine(b, opts, 8, 250, 1000 * 100)
}

func Benchmark_PartitionedPut_Len250_100kRows_Partition16_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedPutGoRoutine(b, opts, 16, 250, 1000 * 100)
}


// Partitioned Put with 100k Rows with length 250 bytes, GoRoutine not-used
func Benchmark_PartitionedPut_Len250_100kRows_Partition1_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedPutNoGoRoutine(b, opts, 1, 250, 1000 * 100)
}

func Benchmark_PartitionedPut_Len250_100kRows_Partition2_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedPutNoGoRoutine(b, opts, 2, 250, 1000 * 100)
}

func Benchmark_PartitionedPut_Len250_100kRows_Partition4_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedPutNoGoRoutine(b, opts, 4, 250, 1000 * 100)
}

func Benchmark_PartitionedPut_Len250_100kRows_Partition8_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedPutNoGoRoutine(b, opts, 8, 250, 1000 * 100)
}

func Benchmark_PartitionedPut_Len250_100kRows_Partition16_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedPutNoGoRoutine(b, opts, 16, 250, 1000 * 100)
}

func Benchmark_PartitionedPut_Len250_100kRows_Partition32_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedPutNoGoRoutine(b, opts, 32, 250, 1000 * 100)
}








///////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////// PARTITIONED GET TESTS BEGINNING /////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////


func benchmarkPartitionedGetNoGoRotine(b *testing.B, opts *opt.Options, numPartitions, valueLength, numInsertions int, readType func(int, int) int) {
	b.StopTimer()
	dirs := genDirs(b, numPartitions)
	defer removeDirs(dirs)

	databases := genDatabases(b, dirs, opts)
	defer closeDBs(databases)


	for i:=0; i<b.N; i++ {
		b.StopTimer()

		keys, values := genKeysAndValues(valueLength, numInsertions)

		for k := 0; k < numInsertions; k++ {
			if numPartitions == 1 {
				databases[0].Put(keys[k], values[k])
			} else {
				partition := getPartitionForTest(keys, k, numPartitions)
				databases[partition].Put(keys[k], values[k])
			}
		}

		b.StartTimer()
		for k := 0; k < numInsertions; k++ {
			keyPos := readType(k, numInsertions)
			if keyPos >= len(keys) {
				b.Fatal("index out of range", keyPos)
			}
			if numPartitions == 1 {
				databases[0].Get(keys[keyPos])
			} else {
				partition := getPartitionForTest(keys, keyPos, numPartitions)
				databases[partition].Get(keys[keyPos])
			}
		}
	}
}


func benchmarkPartitionedGetGoRoutine(b *testing.B, opts *opt.Options, numPartitions, valueLength, numInsertions int, readType func(int, int) int) {
	b.StopTimer()
	dirs := genDirs(b, numPartitions)
	defer removeDirs(dirs)

	databases := genDatabases(b, dirs, opts)
	defer closeDBs(databases)


	for i:=0; i<b.N; i++ {
		b.StopTimer()

		keys, values := genKeysAndValues(valueLength, numInsertions)

		for k := 0; k < numInsertions; k++ {
			if numPartitions == 1 {
				databases[0].Put(keys[k], values[k])
			} else {
				partition := getPartitionForTest(keys, k, numPartitions)
				databases[partition].Put(keys[k], values[k])
			}
		}

		b.StartTimer()
		var wg sync.WaitGroup
		wg.Add(numInsertions)
		for k := 0; k < numInsertions; k++ {
			keyPos := readType(k, numInsertions)
			if keyPos >= len(keys) {
				b.Fatal("index out of range", keyPos)
			}

			go func(kPos int) {
				defer wg.Done()
				if numPartitions == 1 {
					databases[0].Get(keys[kPos])
				} else {
					partition := getPartitionForTest(keys, kPos, numPartitions)
					databases[partition].Get(keys[kPos])
				}
			} (keyPos)

		}
		wg.Wait()
	}
}


func Benchmark_Partitioned_RandomGet_Len250_100kRows_Partition1_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetGoRoutine(b, opts, 1, 250, 1000 * 100, randomRead)
}

func Benchmark_Partitioned_RandomGet_Len250_100kRows_Partition2_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetGoRoutine(b, opts, 2, 250, 1000 * 100, randomRead)
}

func Benchmark_Partitioned_RandomGet_Len250_100kRows_Partition4_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetGoRoutine(b, opts, 4, 250, 1000 * 100, randomRead)
}

func Benchmark_Partitioned_RandomGet_Len250_100kRows_Partition8_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetGoRoutine(b, opts, 8, 250, 1000 * 100, randomRead)
}

func Benchmark_Partitioned_RandomGet_Len250_100kRows_Partition16_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetGoRoutine(b, opts, 16, 250, 1000 * 100, randomRead)
}


func Benchmark_Partitioned_RandomGet_Len250_100kRows_Partition1_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetNoGoRotine(b, opts, 1, 250, 1000 * 100, randomRead)
}

func Benchmark_Partitioned_RandomGet_Len250_100kRows_Partition2_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetNoGoRotine(b, opts, 2, 250, 1000 * 100, randomRead)
}

func Benchmark_Partitioned_RandomGet_Len250_100kRows_Partition4_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetNoGoRotine(b, opts, 4, 250, 1000 * 100, randomRead)
}

func Benchmark_Partitioned_RandomGet_Len250_100kRows_Partition8_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetNoGoRotine(b, opts, 8, 250, 1000 * 100, randomRead)
}

func Benchmark_Partitioned_RandomGet_Len250_100kRows_Partition16_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetNoGoRotine(b, opts, 16, 250, 1000 * 100, randomRead)
}




func Benchmark_Partitioned_ZipfGet_Len250_100kRows_Partition1_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetGoRoutine(b, opts, 1, 250, 1000 * 100, zipfRead)
}

func Benchmark_Partitioned_ZipfGet_Len250_100kRows_Partition2_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetGoRoutine(b, opts, 2, 250, 1000 * 100, zipfRead)
}

func Benchmark_Partitioned_ZipfGet_Len250_100kRows_Partition4_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetGoRoutine(b, opts, 4, 250, 1000 * 100, zipfRead)
}

func Benchmark_Partitioned_ZipfGet_Len250_100kRows_Partition8_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetGoRoutine(b, opts, 8, 250, 1000 * 100, zipfRead)
}

func Benchmark_Partitioned_ZipfGet_Len250_100kRows_Partition16_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetGoRoutine(b, opts, 16, 250, 1000 * 100, zipfRead)
}


func Benchmark_Partitioned_ZipfGet_Len250_100kRows_Partition1_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetNoGoRotine(b, opts, 1, 250, 1000 * 100, zipfRead)
}

func Benchmark_Partitioned_ZipfGet_Len250_100kRows_Partition2_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetNoGoRotine(b, opts, 2, 250, 1000 * 100, zipfRead)
}

func Benchmark_Partitioned_ZipfGet_Len250_100kRows_Partition4_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetNoGoRotine(b, opts, 4, 250, 1000 * 100, zipfRead)
}

func Benchmark_Partitioned_ZipfGet_Len250_100kRows_Partition8_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetNoGoRotine(b, opts, 8, 250, 1000 * 100, zipfRead)
}

func Benchmark_Partitioned_ZipfGet_Len250_100kRows_Partition16_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkPartitionedGetNoGoRotine(b, opts, 16, 250, 1000 * 100, zipfRead)
}




///////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////// Batch Insertion Tests Beginning /////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////



//func Benchmark_Batch_By_BatchSize_1k_250bytes(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkBatch(b, opts, 250, 1000 * 1)
//}
//
//func Benchmark_Batch_By_BatchSize_2k_250bytes(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkBatch(b, opts, 250, 1000 * 2)
//}
//
//func Benchmark_Batch_By_BatchSize_4k_250bytes(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkBatch(b, opts, 250, 1000 * 4)
//}
//
//func Benchmark_Batch_By_BatchSize_8k_250bytes(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkBatch(b, opts, 250, 1000 * 8)
//}
//
//func Benchmark_Batch_By_BatchSize_16k_250bytes(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkBatch(b, opts, 250, 1000 * 16)
//}
//
//func Benchmark_Batch_By_BatchSize_32k_250bytes(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkBatch(b, opts, 250, 1000 * 32)
//}
//
//func Benchmark_Batch_By_BatchSize_64k_250bytes(b *testing.B) {
//	opts := getKlayLDBOptions()
//	benchmarkBatch(b, opts, 250, 1000 * 64)
//}





func Benchmark_Batch_By_BatchSize_1k_500bytes(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatch(b, opts, 500, 1000 * 1)
}

func Benchmark_Batch_By_BatchSize_2k_500bytes(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatch(b, opts, 500, 1000 * 2)
}

func Benchmark_Batch_By_BatchSize_4k_500bytes(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatch(b, opts, 500, 1000 * 4)
}

func Benchmark_Batch_By_BatchSize_8k_500bytes(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatch(b, opts, 500, 1000 * 8)
}

func Benchmark_Batch_By_BatchSize_16k_500bytes(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatch(b, opts, 500, 1000 * 16)
}

func Benchmark_Batch_By_BatchSize_32k_500bytes(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatch(b, opts, 500, 1000 * 32)
}

func Benchmark_Batch_By_BatchSize_64k_500bytes(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatch(b, opts, 500, 1000 * 64)
}


// key = 32 bytes, value = valueLength bytes (100, 250, 500)
func benchmarkBatch(b *testing.B, opts *opt.Options, valueLength, numInsertions int) {
	b.StopTimer()
	dir := genTempDirForTestDB(b)
	defer os.RemoveAll(dir)

	db, err := NewLDBDatabaseWithOptions(dir, opts)
	require.NoError(b, err)
	defer db.Close()



	for i := 0; i < b.N ; i++ {
		b.StopTimer()
		keys, values := genKeysAndValues(valueLength, numInsertions)
		b.StartTimer()
		batch := db.NewBatch()
		for k := 0; k < numInsertions; k++ {
			batch.Put(keys[k], values[k])
		}
		batch.Write()
	}
}

// key = 32 bytes, value = valueLength bytes (100, 250, 500)
func benchmarkBatchPartitionGoRoutine(b *testing.B, opts *opt.Options, valueLength, numInsertions, numPartitions int) {
	b.StopTimer()
	dirs := genDirs(b, numPartitions)
	defer removeDirs(dirs)

	databases := genDatabases(b, dirs, opts)
	defer closeDBs(databases)


	zeroSizeBatch := 0
	batchSizeSum := 0
	numBatches := 0
	for i:=0; i<b.N; i++ {
		b.StopTimer()
		// make same number of batches as numPartitions
		batches := make([]Batch, numPartitions, numPartitions)
		for k:=0; k < numPartitions; k++ {
			batches[k] = databases[k].NewBatch()
		}
		keys, values := genKeysAndValues(valueLength, numInsertions)
		b.StartTimer()
		for k:=0; k < numInsertions; k++ {
			partition := getPartitionForTest(keys, k, numPartitions)
			batches[partition].Put(keys[k], values[k])
		}

		for _, batch := range batches {
			if batch.ValueSize() == 0 {
				zeroSizeBatch++
			}
			batchSizeSum += batch.ValueSize()
			numBatches++
		}
		var wait sync.WaitGroup
		wait.Add(numPartitions)
		for _, batch := range batches {
			go func() {
				defer wait.Done()
				batch.Write()
			}()
		}
		wait.Wait()
	}

	if zeroSizeBatch != 0 {
		b.Log("zeroSizeBatch: ", zeroSizeBatch)
	}
}

func Benchmark_GoRoutine_Overhead_GoRoutine_Disabled(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatch(b, opts, 250,1000 * 100)
}

func Benchmark_GoRoutine_Overhead_GoRoutine_Enabled(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatchPartitionGoRoutine(b, opts, 250, 1000 * 100, 1)
}



func Benchmark_Batch_Partitioned_Len250_100kRows_Partition1_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatchPartitionGoRoutine(b, opts, 250, 1000*100, 1)
}

func Benchmark_Batch_Partitioned_Len250_100kRows_Partition2_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatchPartitionGoRoutine(b, opts, 250, 1000*100, 2)
}

func Benchmark_Batch_Partitioned_Len250_100kRows_Partition4_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatchPartitionGoRoutine(b, opts, 250, 1000*100, 4)
}

func Benchmark_Batch_Partitioned_Len250_100kRows_Partition8_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatchPartitionGoRoutine(b, opts, 250, 1000*100, 8)
}

func Benchmark_Batch_Partitioned_Len250_100kRows_Partition16_GoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatchPartitionGoRoutine(b, opts, 250, 1000*100, 16)
}


func Benchmark_Batch_Partitioned_Len250_100kRows_Partition1_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatchPartitionGoRoutine(b, opts, 250, 1000*100, 1)
}

func Benchmark_Batch_Partitioned_Len250_100kRows_Partition2_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatchPartitionGoRoutine(b, opts, 250, 1000*100, 2)
}

func Benchmark_Batch_Partitioned_Len250_100kRows_Partition4_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatchPartitionGoRoutine(b, opts, 250, 1000*100, 4)
}

func Benchmark_Batch_Partitioned_Len250_100kRows_Partition8_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatchPartitionGoRoutine(b, opts, 250, 1000*100, 8)
}

func Benchmark_Batch_Partitioned_Len250_100kRows_Partition16_NoGoRoutine(b *testing.B) {
	opts := getKlayLDBOptions()
	benchmarkBatchPartitionGoRoutine(b, opts, 250, 1000*100, 16)
}

///////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////// Batch Insertion Tests End /////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////




const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randStrBytes(n int) []byte {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return b
}

func getPartitionForTest(keys [][]byte, index, numPartitions int) int64 {

	return int64(index % numPartitions)
	// TODO-KLAY: CHANGE BELOW LOGIC FROM ROUND-ROBIN TO USE getPartitionForTest
	//key := keys[index]
	//hashString := strings.TrimPrefix(common.Bytes2Hex(key),"0x")
	//if len(hashString) > 15 {
	//	hashString = hashString[:15]
	//}
	//seed, _ := strconv.ParseInt(hashString, 16, 64)
	//partition := seed % int64(numPartitions)
	//
	//return partition
}
