package builder

import (
	"fmt"
	"strings"

	"github.com/go-xorm/core"
	"github.com/pkg/errors"
)

func (b *Builder) limitWriteTo(w Writer) error {
	if !(b.optype == selectType || b.optype == unionType) {
		return errors.New("LIMIT is limited in SELECT and UNION")
	}

	if b.limitation != nil {
		limit := b.limitation
		if limit.offset < 0 || limit.limitN <= 0 {
			return errors.New("Unexpected offset/limitN, set a negative one?")
		}

		selects := b.selects
		// erase limit condition
		b.limitation = nil
		// flush writer, both buffer & args
		ow := w.(*BytesWriter)
		ow.writer.Reset()
		ow.args = nil

		var final *Builder

		switch strings.ToLower(strings.TrimSpace(limit.style)) {
		case core.ORACLE:
			b.selects = append(selects, "ROWNUM RN")
			if limit.offset == 0 {
				final = Select(selects...).From("at", b).
					Where(Lte{"at.ROWNUM": limit.limitN})
			} else {
				sub := Select(append(selects, "RN")...).From("at", b).
					Where(Lte{"at.ROWNUM": limit.offset + limit.limitN})

				if len(selects) == 0 {
					return ErrNotSupportType
				}

				final = Select(selects...).From("att", sub).
					Where(Gt{"att.RN": limit.offset})
			}

			return final.WriteTo(ow)
		case core.SQLITE, core.MYSQL, core.POSTGRES:
			b.WriteTo(ow)

			if limit.offset == 0 {
				fmt.Fprint(ow, " LIMIT ", limit.limitN)
			} else {
				fmt.Fprintf(ow, " LIMIT %v OFFSET %v", limit.limitN, limit.offset)
			}
		case core.MSSQL:
			if limit.offset == 0 {
				if len(selects) == 0 {
					selects = append(selects, "*")
				}

				final = Select(fmt.Sprintf("TOP %d %v", limit.limitN, strings.Join(selects, ","))).
					From("", b).SetNestedFlag(true)
			} else {
				if strings.TrimSpace(limit.pk) == "" {
					return errors.New("Please assign a PK for MsSQL LIMIT operation")
				} else {
					sub := Select(fmt.Sprintf("TOP %d %v", limit.limitN+limit.offset,
						strings.Join(append(selects, limit.pk), ","))).From("", b).SetNestedFlag(true)

					if len(selects) == 0 {
						return ErrNotSupportType
					}

					final = Select(fmt.Sprintf("TOP %d %v", limit.limitN, strings.Join(selects, ","))).
						From("", sub).SetNestedFlag(true).Where(b.cond.And(NotIn(limit.pk, sub)))
				}
			}

			return final.WriteTo(ow)
		default:
			return ErrNotSupportType
		}
	}

	return nil
}