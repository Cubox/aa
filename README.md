# aa
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FCubox-%2Faa.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FCubox-%2Faa?ref=badge_shield)

Small fun programming language, powered by LLVM

## How to compile and use?
You need the llvm Go builtins, go read https://llvm.org/svn/llvm-project/llvm/trunk/bindings/go/README.txt

Then do make and ./aa tests/test.aa

Then do like llc tests/test.aas -o test.s

Then clang test.s -c

Then clang test.o libaa/libaa.c -o test

Then ./test


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FCubox-%2Faa.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FCubox-%2Faa?ref=badge_large)