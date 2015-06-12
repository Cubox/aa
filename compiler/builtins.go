package main

import "llvm.org/llvm/bindings/go/llvm"
import "log"

var (
    builtins = make(map[string]Builtin)
)

type Builtin struct {
    Name string
    Args []ArgumentExprAST
    Body BuiltinBody
    Generate bool
}

type BuiltinBody struct {
    Gen func() llvm.Value
}

func (a BuiltinBody) Codegen() llvm.Value {
    return a.Gen()
}

func init() {
    builtins["itod"] = Builtin {
        Name: "itod",
        Args: []ArgumentExprAST{
            {Name: "i", Type: llvm.Int64Type()},
        },
        Body: BuiltinBody{Gen: Iotd},
        Generate: false,
    }
}

func generateBuiltins() {
    for k, e := range builtins {
        log.Println("Builtin: ", e)
        if builtins[k].Generate {
            log.Println("Generating it!")
            f := module.NamedFunction(e.Name)
            if f.IsNil() {
                log.Println("Generating builtin:", e.Name)
                FunctionAST{Name: e.Name, Args: e.Args, Body: e.Body}.Codegen()
            }
        }
    }
}

func Iotd() llvm.Value {
    return builder.CreateSIToFP(ArgumentExprAST{Name: "i"}.Codegen(), llvm.DoubleType(), "casttmp")
}
