aa:
	cd compiler; go build -x -o ../aa -ldflags "-r \"$$GOPATH/src/llvm.org/llvm/bindings/go/llvm/workdir/llvm_build/lib/\"" ;cd ../

clean:
	rm -f aa

.PHONY: clean aa
