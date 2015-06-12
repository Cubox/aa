package main

import "strconv"
import "os"
import scan "text/scanner"
import "unicode/utf8"
import "log"
import "fmt"
import "llvm.org/llvm/bindings/go/llvm"
import "strings"

var (
    operators = map[rune]int {
        '<': 10,
        '>': 10,
        '+': 20,
        '-': 20,
        '*': 40,
        '/': 40,
        '%': 40,
    }
)

var (
	keywords = make(map[string]func() ExprAST)
)

var (
    fail              = false
    s                 scan.Scanner
    tok               rune
    functionArguments map[string]bool
)

type ExprAST interface {
    String() string
    Codegen() llvm.Value
}

type IntExprAST struct {
    Value int
}

type FloatExprAST struct {
    Value float64
}

type StringExprAST struct {
    Value string
}

type CharExprAST struct {
    Value rune
}

type BinaryExprAST struct {
    Op    rune
    Left  ExprAST
    Right ExprAST
}

type ArgumentExprAST struct {
    Name string
    Type llvm.Type
}

type CallExprAST struct {
    Callee string
    Args   []ExprAST
}

type FunctionAST struct {
    Name string
    Args []ArgumentExprAST
    Body ExprAST
}

type IfExprAST struct {
	Cond, Then, Else ExprAST
}

type ErrorAST struct {
    err string
}

func init() {
	keywords["if"] = parseIf
	keywords["else"] = nil
	keywords["then"] = nil
}

func isKeyword() bool {
	_, ok := keywords[s.TokenText()]
	return ok
}

func tokPrec() int {
    prec, ok := operators[tok]
    if !ok {
        return -1
    } else {
        return prec
    }
}

func isError(ast ExprAST) bool {
    _, ok := ast.(ErrorAST)
    return ok
}

func Error(err string) ExprAST {
    fmt.Println(err, "at", s.Pos(), ":", tok, s.TokenText())
    fail = true
    panic(err)

    return ErrorAST{err: err}
}

func parseChar(char rune) ExprAST {
    tok = s.Scan()
    return CharExprAST{Value: char}
}

func parseString(str string) ExprAST {
    tok = s.Scan()
    return StringExprAST{Value: str}
}

func parseNumber(number string) ExprAST {
    var AST ExprAST

    if tok == scan.Int {
        n, err := strconv.Atoi(number)
        if err != nil {
            log.Println(err)
            fail = true
        }

        AST = IntExprAST{Value: n}
    } else if tok == scan.Float {
        n, err := strconv.ParseFloat(number, 64)
        if err != nil {
            log.Println(err)
            fail = true
        }

        AST = FloatExprAST{Value: n}
    }

    log.Println("Parsed numbed:", AST)

    tok = s.Scan()

    return AST
}

func parseIf() ExprAST {
	tok = s.Scan()

	cond := parseExpression(true)
	if isError(cond) {
		return cond
	}

	log.Println("cond", cond)

	if s.TokenText() != "then" {
		return Error("Excpected then")
	}
	tok = s.Scan()

	then := parseExpression(true)
	if isError(then) {
		return then
	}

	log.Println("then", then)

	if s.TokenText() != "else" {
		return Error("Excpected else")
	}
	tok = s.Scan()

	elseB := parseExpression(true)
	if isError(elseB) {
		return elseB
	}

	log.Println("else", elseB)

	return IfExprAST{Cond: cond, Then: then, Else: elseB}
}

func parseIdent(top bool) ExprAST {
    name := s.TokenText()

    log.Println("Parsed ident:", name, top)

	if function, ok := keywords[name]; ok {
		return function()
	}

    var args []ExprAST

    tok = s.Scan()

    for (tok < 0 || tok == '(') && tok != scan.EOF && top && !isKeyword() { // function call
        log.Println("Parsing an argument (function call):", s.TokenText())
        arg := parseExpression(false)
        if isError(arg) {
            return arg
        }

        args = append(args, arg)
    }

    if tok == '=' && top { // We are defining a function
        arguments := make([]ArgumentExprAST, len(args))
        for i := range args { // Converting to actual arguments
            call, ok := args[i].(CallExprAST)
            if !ok || len(call.Args) != 0 {
                return Error("Caught an invalid argument")
            }

            arguments[i].Name = call.Callee
            lastUS := strings.LastIndex(arguments[i].Name, "_")
            if lastUS > 0 { // Trying to get explicit type
                success := true
                switch arguments[i].Name[lastUS+1:] {
                case "int":
                    arguments[i].Type = llvm.Int64Type()
                case "float":
                    arguments[i].Type = llvm.DoubleType()
                case "chr":
                    arguments[i].Type = llvm.Int8Type()
                default:
                    success = false
                }

                if success {
                    arguments[i].Name = arguments[i].Name[:lastUS]
                } else {
                    arguments[i].Type = llvm.Int64Type()
                }

            } else {
                arguments[i].Type = llvm.Int64Type()
            }
        }

        functionArguments = make(map[string]bool)
        defer func() {
            functionArguments = nil
        }()

        for i := range arguments {
            if functionArguments[arguments[i].Name] {
                return Error("Arguments have the same name!: " + arguments[i].Name)
            }
            functionArguments[arguments[i].Name] = true
        }

        tok = s.Scan()
        log.Println("Trying to parse a function definition, name:", name, "start of body:", s.TokenText())

        body := parseExpression(true)
        if isError(body) {
            return body
        }

        return FunctionAST{Name: name, Args: arguments, Body: body}

    } else if functionArguments != nil && functionArguments[name] {
        // We are an argument from earlier
        log.Println("Argument:", name)
        return ArgumentExprAST{Name: name}
    } else { // Regular call
        log.Println("Regular call:", name)

		builtin, ok := builtins[name]
		if ok {
			log.Println("Setting generate to:", name)
			builtin.Generate = true
			builtins[name] = builtin
		}

        return CallExprAST{Callee: name, Args: args}
    }
}

func parseParenExpr() ExprAST {
    tok = s.Scan()

    log.Println("Parsing parenthesis")

    content := parseExpression(true)

    if isError(content) {
        return content
    }

    if tok != ')' {
        return Error("no ) ? :(")
    }

    log.Println("Done parsing parenthesis")

    tok = s.Scan()

    return content
}

func parsePrimary(top bool) ExprAST {
    text := s.TokenText()

    log.Println("Parsing primary")
    //defer log.Println("Done parsing primary")

    switch tok {
    case scan.Int, scan.Float:
        return parseNumber(text)

    case scan.Char:
        r, _ := utf8.DecodeRuneInString(text)
        return parseChar(r)

    case scan.String, scan.RawString:
        return parseString(text)

    case '(':
        return parseParenExpr()

    case '\n':
        tok = s.Scan()
        return parsePrimary(top)

    case scan.Ident:
        return parseIdent(top)

    case scan.EOF:
        return Error("EOF")

    default:
        return Error("Unknown token")
    }
}

func parseBinOpRHS(prec int, left ExprAST, top bool) ExprAST {
    log.Println("Parsing binary op")
    defer log.Println("Done parsing binary op")

    for {
        log.Println("peek:", s.Peek())
        oldPrec := tokPrec()
        if oldPrec < prec {
            return left
        }

        op := tok
        tok = s.Scan()

        right := parsePrimary(false)

        if isError(right) {
            return right
        }

        newPrec := tokPrec()
        if oldPrec < newPrec {
            right = parseBinOpRHS(oldPrec+1, right, false)
        }

        left = BinaryExprAST{Op: op, Left: left, Right: right}
    }
}

func parseExpression(top bool) ExprAST {
    left := parsePrimary(top)
    if isError(left) {
        return left
    }

    return parseBinOpRHS(0, left, top)
}

func buildAST(list *[]ExprAST, file *os.File) {
    s.Init(file)
    s.Whitespace = 1<<'\t' | 1<<'\r' | 1<<' '
    tok = s.Scan()

    for {
        var AST ExprAST

        switch tok {
        case scan.EOF:
            return

        case '\n':
            tok = s.Scan()
            continue

        default:
            AST = parseExpression(true)
        }

        if isError(AST) {
            fail = true
            break
        }

        *list = append(*list, AST)

        tok = s.Scan()
    }

    if fail {
        fmt.Println("Error caught.")
        os.Exit(1)
    }
}
