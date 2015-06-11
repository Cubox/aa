package main

import "log"
import "io/ioutil"
import "flag"
import "fmt"
import "llvm.org/llvm/bindings/go/llvm"
import "os"

var (
    debug bool
)

func main() {
    flag.BoolVar(&debug, "d", false, "To enable debug printing")
    flag.Parse()

    if debug {
        log.SetFlags(0)
    } else {
        log.SetOutput(ioutil.Discard)
    }

    file, err := os.Open(flag.Arg(0))
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    var list []ExprAST
    buildAST(&list, file)
    for _, elem := range list {
        //fmt.Println(elem)
        elem.Codegen()
    }

    file, err = os.Create(flag.Arg(0) + "s")
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    module.Dump()
    llvm.VerifyModule(module, llvm.AbortProcessAction)
    llvm.WriteBitcodeToFile(module, file)
}
