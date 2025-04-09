package parser

import (
    "bytes"
    "fmt"
    "encoding/hex"

    "github.com/cedar-policy/cedar-go/internal/schema/token"
)

func (l *Lexer) lex() (pos token.Position, tok token.Type, lit string, err error) {
    for {
        lit = ""
        l.pos.Offset = l.cursor
        l.pos.Column = l.cursor - l.lineStart + 1
        l.token = l.cursor
        pos = l.pos

        /*!re2c
        re2c:yyfill:enable = 0;
        re2c:flags:nested-ifs = 1;
        re2c:define:YYCTYPE = byte;
        re2c:define:YYPEEK = "l.input[l.cursor]";
        re2c:define:YYSKIP = "l.cursor += 1";

        end = [\x00];
        end { l.cursor -= 1; tok = token.EOF; return }
        * { err = ErrUnrecognizedToken; return }

        // Whitespace and new lines
        eol = ("\r\n" | "\n");
        eol {
            l.pos.Line += 1
            l.pos.Column = 1
            l.lineStart = l.cursor
            continue
        }

            // Skip whitespace
        [ \t]+ {
            continue
        }

        // Comments
        "//" [^\r\n\x00]* { tok = token.COMMENT; lit = l.literal(); return }

        "namespace" { tok = token.NAMESPACE; lit = "namespace"; return }
        "entity" { tok = token.ENTITY; lit = "entity"; return }
        "action" { tok = token.ACTION; lit = "action"; return }
        "type" { tok = token.TYPE; lit = "type"; return }
        "in" { tok = token.IN; lit = "in"; return }
        "tags" { tok = token.TAGS; lit = "tags"; return }
        "appliesTo" { tok = token.APPLIES_TO; lit = "appliesTo"; return }
        "principal" { tok = token.PRINCIPAL; lit = "principal"; return }
        "resource" { tok = token.RESOURCE; lit = "resource"; return }
        "context" { tok = token.CONTEXT; lit = "context"; return }

        // Operators and punctuation
        "{" { tok = token.LEFTBRACE; lit = "{"; return }
        "}" { tok = token.RIGHTBRACE; lit = "}"; return }
        "[" { tok = token.LEFTBRACKET; lit = "["; return }
        "]" { tok = token.RIGHTBRACKET; lit = "]"; return }
        "<" { tok = token.LEFTANGLE; lit = "<"; return }
        ">" { tok = token.RIGHTANGLE; lit = ">"; return }
        ":" { tok = token.COLON; lit = ":"; return }
        ";" { tok = token.SEMICOLON; lit = ";"; return }
        "," { tok = token.COMMA; lit = ","; return }
        "=" { tok = token.EQUALS; lit = "="; return }
        "?" { tok = token.QUESTION; lit = "?"; return }
        "::" { tok = token.DOUBLECOLON; lit = "::"; return }

        // Strings
        ["] { return l.lexString('"') }

        // Identifiers
        id = [a-zA-Z_][a-zA-Z_0-9]*;
        id { tok = token.IDENT; lit = l.literal(); return }
        */
    }
}

func (l *Lexer) lexString(quote byte) (pos token.Position, tok token.Type, lit string, err error) {
    pos = l.pos
    marker := 0
    var buf bytes.Buffer
    buf.WriteByte(quote)
    for {
        var u byte

        /*!re2c
        re2c:yyfill:enable = 0;
        re2c:flags:nested-ifs = 1;
        re2c:define:YYBACKUP = "marker = l.cursor";
        re2c:define:YYRESTORE = "l.cursor = marker";
        re2c:define:YYPEEK = "l.input[l.cursor]";
        re2c:define:YYSKIP = "l.cursor += 1";

        * { err = ErrInvalidString; return }
        [\x00] {
            l.cursor -= 1 // make sure we don't overflow next lex call
            err = ErrUnterminatedString
            tok = token.EOF
            pos = l.pos
            return
        }
        [^\n\\] {
            u = yych
            buf.WriteByte(u)
            if u == quote {
                tok = token.STRING
                pos = l.pos
                lit = string(buf.Bytes())
                return
            }
            continue
        }
        // Unicode escape sequences
        "\\u{" [0-9A-Fa-f]+ "}" {
            // Handle the hex digits between the braces
            hexStr := string(l.input[marker+2:l.cursor-1])  // Strip off \u{ and }
            if len(hexStr) % 2 != 0 {
                hexStr = "0" + hexStr
            }
            var val []byte
            val, err = hex.DecodeString(hexStr)
            if err != nil {
                pos = l.pos
                lit = string(buf.Bytes())
                err = fmt.Errorf("%w: %s", ErrInvalidString, err)
                return
            }
            buf.Write(val)
            continue
        }
        "\\0"  { buf.WriteByte(0); continue }
        "\\n"  { buf.WriteByte('\n'); continue }
        "\\r"  { buf.WriteByte('\r'); continue }
        "\\t"  { buf.WriteByte('\t'); continue }
        "\\\\" { buf.WriteByte('\\'); continue }
        "\\'"  { buf.WriteByte('\''); continue }
        "\\\"" { buf.WriteByte('"'); continue }
        */
    }
}
