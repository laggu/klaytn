pragma solidity ^0.4.24;

/**
 * @title AddressBook
 */
contract AddressBook {

    bool isInitialized = false;

    address[] adminList;
    uint requirement;

    address[] cnNodeIdList;
    address[] cnStakingContractList;
    address[] cnRewardAddressList;
    address pocContract;
    address kirContract;


    //@TODO multi-sig function modifier
    modifier multiSig() {
      _;
    }

    modifier afterInit() {
      require(isInitialized == true);
      _;
    }


    function init (address[] _adminList, address[] _cnNodeIdList, address[] _cnStakingContractList, address[] _cnRewardAddressList,
      address _pocContract, address _kirContract) public {
      require(msg.sender == 0x0);
      requirement = 1;

      adminList = _adminList;
      cnNodeIdList = _cnNodeIdList;
      cnStakingContractList = _cnStakingContractList;
      cnRewardAddressList = _cnRewardAddressList;
      pocContract = _pocContract;
      kirContract = _kirContract;
    }


    //TEST용 init 함수
    function initTest () public {
      requirement = 1;

      adminList.push(0x18254160af9c10f43db77566c294db7d8182caca);

      //> personal.importRawKey("1696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
      // "0x9b0bf94f5ab62c73e454dfc55adb2d2fa6cd3af5"
      // > personal.importRawKey("2696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
      // "0x4d83f1795ecdd684e94f1c5893ae6904ebeaeb94"
      //> personal.importRawKey("3696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
      // "0x9b58fe24f7a7cb9d102e21b3376bd80eefdc320b"
      // > personal.importRawKey("4696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
      // "0xe60bf7b625e54e9f67767fad0e564f6aec297652"
      // > personal.importRawKey("5696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
      // "0xac5e047d39692be8c81d0724543d5de721d0dd54"
      // > personal.importRawKey("6696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
      // "0x382ef85439dc874a0f55ab4d9801a5056e371b37"
      // > personal.importRawKey("7696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
      // "0xb821a659c21cb39745144931e71d0e9d09c8647f"
      // > personal.importRawKey("8696515745bfc8769c6152cfa0f6c56bf356692a037c4e3fce31faeb66f65b07", "")
      // "0xc1094c666657937ab7ed23040207a8ee68781350"

      // NodeId of CNs
      cnNodeIdList.push(0x060107a87178fcd01822d616e17eb5b26dd279ad);
      cnNodeIdList.push(0x82ad1aea897ba3e72ee3632c08fe7596a9ff9c3f);
      cnNodeIdList.push(0x0994ebd481f77768a8bf20d35d8f405ce21d2176);
      cnNodeIdList.push(0x78d2ad8c09dce2eba9c25ef8201607ead8637b8e);

      // Staking contracts
      cnStakingContractList.push(0xac5e047d39692be8c81d0724543d5de721d0dd54);
      cnStakingContractList.push(0x382ef85439dc874a0f55ab4d9801a5056e371b37);
      cnStakingContractList.push(0xb821a659c21cb39745144931e71d0e9d09c8647f);
      cnStakingContractList.push(0xc1094c666657937ab7ed23040207a8ee68781350);

      // Reward addresses
      cnRewardAddressList.push(0x9b0bf94f5ab62c73e454dfc55adb2d2fa6cd3af5);
      cnRewardAddressList.push(0x4d83f1795ecdd684e94f1c5893ae6904ebeaeb94);
      cnRewardAddressList.push(0x9b58fe24f7a7cb9d102e21b3376bd80eefdc320b);
      cnRewardAddressList.push(0xe60bf7b625e54e9f67767fad0e564f6aec297652);

      pocContract = 0x142441cB0896D4cD2eCdF3328d8D841A07f4B04f;
      kirContract = 0xA36c921743D63361258FDBD107B906Be0ad87940;
    }


    function completeInitialization() public multiSig returns (bool) {
      require(isInitialized == false);

      require(adminList.length != 0);

      isInitialized = true;
    }


    function testAddCn() public afterInit returns (bool) {
      for (uint i = 0; i < 10; i++) {
        cnNodeIdList.push(0xa22499738b961e56fb833D9368241A6a789E77C4);
        cnStakingContractList.push(0xC63bA6B8a9F2d33eE26a7fD0A59d155131F2701A);
        cnRewardAddressList.push(0x193eFB18f18F93E32e418E383714279bDade4a81);
      }
    }


    function getAllAddressInfo() public afterInit view returns(address[], address[], address[], address, address) {
      return (cnNodeIdList, cnStakingContractList, cnRewardAddressList, pocContract, kirContract);
    }
}
