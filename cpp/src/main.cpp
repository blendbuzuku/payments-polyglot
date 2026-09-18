#include <fstream>
#include <iostream>

#include "processor.h"

int main(int argc, char* argv[]) {
    if (argc < 2) {
        std::cerr << "usage: settle <payments.csv>\n";
        return 1;
    }

    std::ifstream file(argv[1]);  // closed by its destructor, on every path out of main
    if (!file) {
        std::cerr << "cannot open " << argv[1] << '\n';
        return 1;
    }

    settle::BatchProcessor processor;
    processor.run(file, std::cout);
    return 0;
}
