// Code generated by re2go 4.3 on Mon Jul  7 15:42:21 2025, DO NOT EDIT.
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

        
{
	var yych byte
	yych = l.input[l.cursor]
	switch (yych) {
	case 0x00:
		goto yy1
	case '\t':
		fallthrough
	case ' ':
		goto yy4
	case '\n':
		goto yy5
	case '\r':
		goto yy6
	case '"':
		goto yy7
	case '(':
		goto yy8
	case ')':
		goto yy9
	case ',':
		goto yy10
	case '/':
		goto yy11
	case ':':
		goto yy12
	case ';':
		goto yy13
	case '<':
		goto yy14
	case '=':
		goto yy15
	case '>':
		goto yy16
	case '?':
		goto yy17
	case '@':
		goto yy18
	case 'A','B','C','D','E','F','G','H','I','J','K','L','M','N','O','P','Q','R','S','T','U','V','W','X','Y','Z':
		fallthrough
	case '_':
		fallthrough
	case 'b':
		fallthrough
	case 'd':
		fallthrough
	case 'f','g','h':
		fallthrough
	case 'j','k','l','m':
		fallthrough
	case 'o':
		fallthrough
	case 'q':
		fallthrough
	case 's':
		fallthrough
	case 'u','v','w','x','y','z':
		goto yy19
	case '[':
		goto yy22
	case ']':
		goto yy23
	case 'a':
		goto yy24
	case 'c':
		goto yy25
	case 'e':
		goto yy26
	case 'i':
		goto yy27
	case 'n':
		goto yy28
	case 'p':
		goto yy29
	case 'r':
		goto yy30
	case 't':
		goto yy31
	case '{':
		goto yy32
	case '}':
		goto yy33
	default:
		goto yy2
	}
yy1:
	l.cursor += 1
	{ l.cursor -= 1; tok = token.EOF; return }
yy2:
	l.cursor += 1
yy3:
	{ err = ErrUnrecognizedToken; return }
yy4:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == '\t') {
		goto yy4
	}
	if (yych == ' ') {
		goto yy4
	}
	{
            continue
        }
yy5:
	l.cursor += 1
	{
            l.pos.Line += 1
            l.pos.Column = 1
            l.lineStart = l.cursor
            continue
        }
yy6:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == '\n') {
		goto yy5
	}
	goto yy3
yy7:
	l.cursor += 1
	{ return l.lexString('"') }
yy8:
	l.cursor += 1
	{ tok = token.LEFTPAREN; lit = "("; return }
yy9:
	l.cursor += 1
	{ tok = token.RIGHTPAREN; lit = ")"; return }
yy10:
	l.cursor += 1
	{ tok = token.COMMA; lit = ","; return }
yy11:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == '/') {
		goto yy34
	}
	goto yy3
yy12:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == ':') {
		goto yy36
	}
	{ tok = token.COLON; lit = ":"; return }
yy13:
	l.cursor += 1
	{ tok = token.SEMICOLON; lit = ";"; return }
yy14:
	l.cursor += 1
	{ tok = token.LEFTANGLE; lit = "<"; return }
yy15:
	l.cursor += 1
	{ tok = token.EQUALS; lit = "="; return }
yy16:
	l.cursor += 1
	{ tok = token.RIGHTANGLE; lit = ">"; return }
yy17:
	l.cursor += 1
	{ tok = token.QUESTION; lit = "?"; return }
yy18:
	l.cursor += 1
	{ tok = token.AT; lit = "@"; return }
yy19:
	l.cursor += 1
	yych = l.input[l.cursor]
yy20:
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy21
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy21
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy21:
	{ tok = token.IDENT; lit = l.literal(); return }
yy22:
	l.cursor += 1
	{ tok = token.LEFTBRACKET; lit = "["; return }
yy23:
	l.cursor += 1
	{ tok = token.RIGHTBRACKET; lit = "]"; return }
yy24:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'c') {
		goto yy37
	}
	if (yych == 'p') {
		goto yy38
	}
	goto yy20
yy25:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'o') {
		goto yy39
	}
	goto yy20
yy26:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'n') {
		goto yy40
	}
	goto yy20
yy27:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'n') {
		goto yy41
	}
	goto yy20
yy28:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'a') {
		goto yy43
	}
	goto yy20
yy29:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'r') {
		goto yy44
	}
	goto yy20
yy30:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'e') {
		goto yy45
	}
	goto yy20
yy31:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'a') {
		goto yy46
	}
	if (yych == 'y') {
		goto yy47
	}
	goto yy20
yy32:
	l.cursor += 1
	{ tok = token.LEFTBRACE; lit = "{"; return }
yy33:
	l.cursor += 1
	{ tok = token.RIGHTBRACE; lit = "}"; return }
yy34:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= '\n') {
		if (yych <= 0x00) {
			goto yy35
		}
		if (yych <= '\t') {
			goto yy34
		}
	} else {
		if (yych != '\r') {
			goto yy34
		}
	}
yy35:
	{ tok = token.COMMENT; lit = l.literal(); return }
yy36:
	l.cursor += 1
	{ tok = token.DOUBLECOLON; lit = "::"; return }
yy37:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 't') {
		goto yy48
	}
	goto yy20
yy38:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'p') {
		goto yy49
	}
	goto yy20
yy39:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'n') {
		goto yy50
	}
	goto yy20
yy40:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 's') {
		goto yy20
	}
	if (yych <= 't') {
		goto yy51
	}
	if (yych <= 'u') {
		goto yy52
	}
	goto yy20
yy41:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy42
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy42
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy42:
	{ tok = token.IN; lit = "in"; return }
yy43:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'm') {
		goto yy53
	}
	goto yy20
yy44:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'i') {
		goto yy54
	}
	goto yy20
yy45:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 's') {
		goto yy55
	}
	goto yy20
yy46:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'g') {
		goto yy56
	}
	goto yy20
yy47:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'p') {
		goto yy57
	}
	goto yy20
yy48:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'i') {
		goto yy58
	}
	goto yy20
yy49:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'l') {
		goto yy59
	}
	goto yy20
yy50:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 't') {
		goto yy60
	}
	goto yy20
yy51:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'i') {
		goto yy61
	}
	goto yy20
yy52:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'm') {
		goto yy62
	}
	goto yy20
yy53:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'e') {
		goto yy64
	}
	goto yy20
yy54:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'n') {
		goto yy65
	}
	goto yy20
yy55:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'o') {
		goto yy66
	}
	goto yy20
yy56:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 's') {
		goto yy67
	}
	goto yy20
yy57:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'e') {
		goto yy69
	}
	goto yy20
yy58:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'o') {
		goto yy71
	}
	goto yy20
yy59:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'i') {
		goto yy72
	}
	goto yy20
yy60:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'e') {
		goto yy73
	}
	goto yy20
yy61:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 't') {
		goto yy74
	}
	goto yy20
yy62:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy63
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy63
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy63:
	{ tok = token.ENUM; lit = "enum"; return }
yy64:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 's') {
		goto yy75
	}
	goto yy20
yy65:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'c') {
		goto yy76
	}
	goto yy20
yy66:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'u') {
		goto yy77
	}
	goto yy20
yy67:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy68
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy68
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy68:
	{ tok = token.TAGS; lit = "tags"; return }
yy69:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy70
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy70
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy70:
	{ tok = token.TYPE; lit = "type"; return }
yy71:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'n') {
		goto yy78
	}
	goto yy20
yy72:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'e') {
		goto yy80
	}
	goto yy20
yy73:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'x') {
		goto yy81
	}
	goto yy20
yy74:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'y') {
		goto yy82
	}
	goto yy20
yy75:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'p') {
		goto yy84
	}
	goto yy20
yy76:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'i') {
		goto yy85
	}
	goto yy20
yy77:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'r') {
		goto yy86
	}
	goto yy20
yy78:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy79
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy79
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy79:
	{ tok = token.ACTION; lit = "action"; return }
yy80:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 's') {
		goto yy87
	}
	goto yy20
yy81:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 't') {
		goto yy88
	}
	goto yy20
yy82:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy83
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy83
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy83:
	{ tok = token.ENTITY; lit = "entity"; return }
yy84:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'a') {
		goto yy90
	}
	goto yy20
yy85:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'p') {
		goto yy91
	}
	goto yy20
yy86:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'c') {
		goto yy92
	}
	goto yy20
yy87:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'T') {
		goto yy93
	}
	goto yy20
yy88:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy89
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy89
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy89:
	{ tok = token.CONTEXT; lit = "context"; return }
yy90:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'c') {
		goto yy94
	}
	goto yy20
yy91:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'a') {
		goto yy95
	}
	goto yy20
yy92:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'e') {
		goto yy96
	}
	goto yy20
yy93:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'o') {
		goto yy98
	}
	goto yy20
yy94:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'e') {
		goto yy100
	}
	goto yy20
yy95:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == 'l') {
		goto yy102
	}
	goto yy20
yy96:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy97
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy97
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy97:
	{ tok = token.RESOURCE; lit = "resource"; return }
yy98:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy99
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy99
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy99:
	{ tok = token.APPLIESTO; lit = "appliesTo"; return }
yy100:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy101
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy101
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy101:
	{ tok = token.NAMESPACE; lit = "namespace"; return }
yy102:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych <= 'Z') {
		if (yych <= '/') {
			goto yy103
		}
		if (yych <= '9') {
			goto yy19
		}
		if (yych >= 'A') {
			goto yy19
		}
	} else {
		if (yych <= '_') {
			if (yych >= '_') {
				goto yy19
			}
		} else {
			if (yych <= '`') {
				goto yy103
			}
			if (yych <= 'z') {
				goto yy19
			}
		}
	}
yy103:
	{ tok = token.PRINCIPAL; lit = "principal"; return }
}

    }
}

func (l *Lexer) lexString(quote byte) (pos token.Position, tok token.Type, lit string, err error) {
    pos = l.pos
    marker := 0
    var buf bytes.Buffer
    buf.WriteByte(quote)
    for {
        var u byte

        
{
	var yych byte
	yych = l.input[l.cursor]
	if (yych <= '\n') {
		if (yych <= 0x00) {
			goto yy105
		}
		if (yych <= '\t') {
			goto yy106
		}
		goto yy107
	} else {
		if (yych == '\\') {
			goto yy109
		}
		goto yy106
	}
yy105:
	l.cursor += 1
	{
            l.cursor -= 1 // make sure we don't overflow next lex call
            err = ErrUnterminatedString
            tok = token.EOF
            pos = l.pos
            return
        }
yy106:
	l.cursor += 1
	{
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
yy107:
	l.cursor += 1
yy108:
	{ err = ErrInvalidString; return }
yy109:
	l.cursor += 1
	marker = l.cursor
	yych = l.input[l.cursor]
	if (yych <= '\\') {
		if (yych <= '\'') {
			if (yych == '"') {
				goto yy110
			}
			if (yych <= '&') {
				goto yy108
			}
			goto yy111
		} else {
			if (yych == '0') {
				goto yy112
			}
			if (yych <= '[') {
				goto yy108
			}
			goto yy113
		}
	} else {
		if (yych <= 'r') {
			if (yych == 'n') {
				goto yy114
			}
			if (yych <= 'q') {
				goto yy108
			}
			goto yy115
		} else {
			if (yych <= 's') {
				goto yy108
			}
			if (yych <= 't') {
				goto yy116
			}
			if (yych <= 'u') {
				goto yy117
			}
			goto yy108
		}
	}
yy110:
	l.cursor += 1
	{ buf.WriteByte('"'); continue }
yy111:
	l.cursor += 1
	{ buf.WriteByte('\''); continue }
yy112:
	l.cursor += 1
	{ buf.WriteByte(0); continue }
yy113:
	l.cursor += 1
	{ buf.WriteByte('\\'); continue }
yy114:
	l.cursor += 1
	{ buf.WriteByte('\n'); continue }
yy115:
	l.cursor += 1
	{ buf.WriteByte('\r'); continue }
yy116:
	l.cursor += 1
	{ buf.WriteByte('\t'); continue }
yy117:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == '{') {
		goto yy119
	}
yy118:
	l.cursor = marker
	goto yy108
yy119:
	l.cursor += 1
	yych = l.input[l.cursor]
	if (yych == '}') {
		goto yy118
	}
	goto yy121
yy120:
	l.cursor += 1
	yych = l.input[l.cursor]
yy121:
	if (yych <= 'F') {
		if (yych <= '/') {
			goto yy118
		}
		if (yych <= '9') {
			goto yy120
		}
		if (yych <= '@') {
			goto yy118
		}
		goto yy120
	} else {
		if (yych <= 'f') {
			if (yych <= '`') {
				goto yy118
			}
			goto yy120
		} else {
			if (yych != '}') {
				goto yy118
			}
		}
	}
	l.cursor += 1
	{
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
}

    }
}
