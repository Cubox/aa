package main

import "llvm.org/llvm/bindings/go/llvm"
import "fmt"
import "log"
import "os"

var (
    builder  = llvm.NewBuilder()
    module   = llvm.NewModule("aa")
    dryRun   = false
    argsList map[string]llvm.Value
)

func CodegenError(err string) llvm.Value {
    panic(err)
    os.Exit(1)

    return llvm.Value{}
}

func Type(value llvm.Value) llvm.TypeKind {
    t := value.Type()
    tk := t.TypeKind()

    if tk == llvm.FunctionTypeKind { // Function are actually pointers to them
        return t.ReturnType().TypeKind()
    } else {
        return tk
    }
}

func (a ArgumentExprAST) Codegen() llvm.Value {
    arg, ok := argsList[a.Name]
    if ok {
        return arg
    } else {
        fmt.Println(argsList)
        return CodegenError("Argument " + a.Name + " not found.")
    }
}

func (a IntExprAST) Codegen() llvm.Value {
    return llvm.ConstInt(llvm.Int64Type(), uint64(a.Value), true)

}

func (a FloatExprAST) Codegen() llvm.Value {
    return llvm.ConstFloat(llvm.DoubleType(), a.Value)

}

func (a StringExprAST) Codegen() llvm.Value {
    return llvm.ConstString(a.Value, false)
}

func (a CharExprAST) Codegen() llvm.Value {
    return llvm.ConstInt(llvm.Int8Type(), uint64(a.Value), true)
}

func (a BinaryExprAST) Codegen() llvm.Value {
    left := a.Left.Codegen()
    right := a.Right.Codegen()

    t := Type(left)

    if t != Type(right) {
        CodegenError("Different types for binop " + a.Left.String() + " " + string(a.Op) + " " + a.Right.String())
    }

    if t == llvm.IntegerTypeKind {
        switch a.Op {
        case '+':
            return builder.CreateAdd(left, right, "addtmp")
        case '-':
            return builder.CreateSub(left, right, "subtmp")
        case '*':
            return builder.CreateMul(left, right, "multmp")
        case '<':
            cmp := builder.CreateICmp(llvm.IntULT, left, right, "cmptmp")
            return builder.CreateSExt(cmp, llvm.Int8Type(), "booltmp")
        default:
            return CodegenError("Unknown operator: " + string(a.Op))
        }
    } else if t == llvm.DoubleTypeKind {
        switch a.Op {
        case '+':
            return builder.CreateFAdd(left, right, "faddtmp")
        case '-':
            return builder.CreateFSub(left, right, "fsubtmp")
        case '*':
            return builder.CreateFMul(left, right, "fmultmp")
        case '<':
            cmp := builder.CreateFCmp(llvm.FloatULT, left, right, "fcmptmp")
            return builder.CreateSExt(cmp, llvm.Int8Type(), "booltmp")
        default:
            return CodegenError("Unknown operator: " + string(a.Op))
        }
    } else {
        return CodegenError("Using an operator on an invalid type: " + t.String())
    }
}

func (a CallExprAST) Codegen() llvm.Value {
    log.Println("Creating call to:", a)
    function := module.NamedFunction(a.Callee)
    if function.IsNil() || function.IsNull() {
        return CodegenError("Unknown function " + a.Callee)
    }
    if function.ParamsCount() != len(a.Args) {
        return CodegenError("Args number is different than known ones")
    }

    args := make([]llvm.Value, len(a.Args))

    for i := range args {
        args[i] = a.Args[i].Codegen()
    }

    return builder.CreateCall(function, args, "calltmp")
}

func (a FunctionAST) Codegen() llvm.Value {
    log.Println("Generate function:", a.Name)

    if call, ok := a.Body.(CallExprAST); ok { // We need to see if it's an extern
        var t llvm.Type
        switch call.Callee {
        case "int":
            t = llvm.Int64Type()
        case "float":
            t = llvm.DoubleType()
        case "chr":
            t = llvm.Int8Type()
        }
        if !t.IsNil() {
            args := make([]llvm.Type, len(a.Args))
            for i := range args {
                args[i] = a.Args[i].Type
            }
            fType := llvm.FunctionType(t, args, false)
            return llvm.AddFunction(module, a.Name, fType)
        }
    }

    args := make([]llvm.Type, len(a.Args))
    argsList = make(map[string]llvm.Value)

    for i := range args {
        args[i] = a.Args[i].Type
        // First run, we don't care about the value
        argsList[a.Args[i].Name] = llvm.ConstNull(a.Args[i].Type)
    }

    restoreDry := false
    if !dryRun {
        dryRun = true
        restoreDry = true
    }

    builder.ClearInsertionPoint()
    retType := a.Body.Codegen().Type()

    if restoreDry {
        dryRun = false
    }

    fType := llvm.FunctionType(retType, args, false)

    currentModule := module
    if dryRun {
        currentModule = llvm.NewModule("tmp")
    }

    f := llvm.AddFunction(currentModule, a.Name, fType)

    if f.Name() != a.Name {
        oldF := module.NamedFunction(a.Name)
        f.EraseFromParentAsFunction()
        f = oldF
    }

    argsList = make(map[string]llvm.Value)
    for i := range a.Args {
        argsList[a.Args[i].Name] = f.Param(i)
        f.Param(i).SetName(a.Args[i].Name)
    }

    bb := llvm.AddBasicBlock(f, "entry")
    builder.SetInsertPointAtEnd(bb)

    ret := a.Body.Codegen()

    builder.CreateRet(ret)

    if dryRun {
        currentModule.Dispose()
    } else {
        llvm.VerifyFunction(f, llvm.AbortProcessAction)
    }

    return f
}

func (a IfExprAST) Codegen() llvm.Value {
    cond := a.Cond.Codegen()
    var condV llvm.Value

    if Type(cond) == llvm.IntegerTypeKind {
        condV = builder.CreateICmp(llvm.IntNE, cond,
            llvm.ConstInt(llvm.Int8Type(), uint64(0), true), "ifcond")
    } else if Type(cond) == llvm.DoubleTypeKind {
        condV = builder.CreateFCmp(llvm.FloatONE, cond,
            llvm.ConstFloat(llvm.DoubleType(), 0), "ifcond")
    } else {
        return CodegenError("Using an if on an invalid type: " + cond.Type().String())
    }

    if dryRun {
        return a.Then.Codegen()
    }

    curFunc := builder.GetInsertBlock().Parent()
    thenBB := llvm.AddBasicBlock(curFunc, "then")
    elseBB := llvm.AddBasicBlock(curFunc, "else")
    mergeBB := llvm.AddBasicBlock(curFunc, "merge")

    builder.CreateCondBr(condV, thenBB, elseBB)
    builder.SetInsertPointAtEnd(thenBB)

    thenV := a.Then.Codegen()

    builder.CreateBr(mergeBB)
    thenBB = builder.GetInsertBlock()

    elseBB.MoveAfter(thenBB)
    builder.SetInsertPointAtEnd(elseBB)

    elseV := a.Else.Codegen()

    builder.CreateBr(mergeBB)
    elseBB = builder.GetInsertBlock()
    mergeBB.MoveAfter(elseBB)

    builder.SetInsertPointAtEnd(mergeBB)

    phi := builder.CreatePHI(thenV.Type(), "iftmp")

    phi.AddIncoming([]llvm.Value{thenV, elseV}, []llvm.BasicBlock{thenBB, elseBB})
    return phi
}

func (a ErrorAST) Codegen() llvm.Value {
    return CodegenError("Tried to Codegen an ErrorAST")
}
