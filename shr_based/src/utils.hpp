#include "boost/tokenizer.hpp"
#include "FloatingPoint/floating-point.h"
#include "FloatingPoint/fp-math.h"

#include <iostream>
#include <fstream>

#define PUBLIC_ROLE 0
#define ALICE_ROLE 1
#define BOB_ROLE 2


typedef uint8_t u8;
typedef uint32_t u32;
typedef uint64_t u64;

typedef double dataType;

typedef struct DataRowStruct {
    int number;
    dataType data;
} DataRow;

typedef struct ConfigStruct {
    int role;
    int opt;
    int port;
    
    std::string addr;
    std::string csvPth;
    std::string shrPth;
    std::string resPth;
    std::string bPth;
} Config;

enum class Operation { ADD, SUB, MUL, DIV, CHEAP_ADD, CHEAP_DIV };


std::vector<dataType> loadDataFromCsv(std::string& fileName, std::string& basic);
void shareInputData(FPOp *fpOp, int party, std::vector<dataType>& dataVec, std::string& fileName, std::string& basic);

std::pair<FPArray, FPArray> getArrayFromBinary(FPOp *fpOp, int party, std::string& fileName, std::string& basic);
std::vector<dataType> doCalculate(FPOp *fpOp, int party, FPArray& fpArr1, FPArray& fpArr2, Operation op);


void serializeShare(u8 *compS, u8 *compZ, u64 *compM, u64 *compE, size_t dataSize, std::string& fileName) {
    std::ofstream outFile(fileName, std::ios::binary);
    if (!outFile.is_open()) {
        throw std::runtime_error("Cannot open file for writing");
    }

    outFile.write(reinterpret_cast<const char *>(&dataSize), sizeof(dataSize));
    outFile.write(reinterpret_cast<const char *>(compS), dataSize * sizeof(u8));
    outFile.write(reinterpret_cast<const char *>(compZ), dataSize * sizeof(u8));
    outFile.write(reinterpret_cast<const char *>(compM), dataSize * sizeof(u64));
    outFile.write(reinterpret_cast<const char *>(compE), dataSize * sizeof(u64));

    outFile.close();
}


std::tuple<std::vector<u8>, std::vector<u8>, std::vector<u64>, std::vector<u64>, size_t> deserializeShare(std::string& fileName) {
    std::ifstream inFile(fileName, std::ios::binary);
    if (!inFile.is_open()) {
        throw std::runtime_error("Cannot open file for reading");
    }

    size_t dataSize = 0;
    inFile.read(reinterpret_cast<char *>(&dataSize), sizeof(dataSize));

    std::vector<u8> compS(dataSize);
    std::vector<u8> compZ(dataSize);
    std::vector<u64> compM(dataSize);
    std::vector<u64> compE(dataSize);

    inFile.read(reinterpret_cast<char *>(compS.data()), dataSize * sizeof(u8));
    inFile.read(reinterpret_cast<char *>(compZ.data()), dataSize * sizeof(u8));
    inFile.read(reinterpret_cast<char *>(compM.data()), dataSize * sizeof(u64));
    inFile.read(reinterpret_cast<char *>(compE.data()), dataSize * sizeof(u64));

    inFile.close();
    return std::make_tuple(compS, compZ, compM, compE, dataSize);
}


void saveResultToFile(std::vector<dataType>& result, std::string& fileName, std::string& basic) {
    auto filePath = basic + fileName;
    std::ofstream outFile(filePath);
    if (!outFile.is_open()) {
        std::cerr << "Cannot open file for writing" << std::endl;
        return;
    }

    for (const float& value : result) {
        outFile << value << std::endl;
    }

    outFile.close();
}
