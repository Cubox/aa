package main

import "fmt"

func (a IntExprAST) String() string {
    return fmt.Sprintf("Int: %v", a.Value)
}

func (a FloatExprAST) String() string {
    return fmt.Sprintf("Float: %v", a.Value)
}

func (a StringExprAST) String() string {
    return fmt.Sprintf("String: %v", a.Value)
}

func (a CharExprAST) String() string {
    return fmt.Sprintf("Char: %c / %d", a.Value, a.Value)
}

func (a BinaryExprAST) String() string {
    return fmt.Sprintf("Binop: %c Left: %v Right: %v", a.Op, a.Left, a.Right)
}

func (a CallExprAST) String() string {
    return fmt.Sprintf("Call. Name: %s Args: %v", a.Callee, a.Args)
}

func (a FunctionAST) String() string {
    return fmt.Sprintf("Function: Name: %s Arguments: %v Body: %v", a.Name, a.Args, a.Body)
}

func (a ErrorAST) String() string {
    return fmt.Sprintf("Error: %s", a.err)
}

func (a ArgumentExprAST) String() string {
    return fmt.Sprintf("Argument: %s", a.Name)
}
