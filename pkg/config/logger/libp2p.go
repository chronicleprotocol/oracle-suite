//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package logger

import (
	"fmt"
	"math"
	"time"

	libp2pLog "github.com/ipfs/go-log/v2"
	"go.uber.org/zap/zapcore"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

func EnableLibP2PLogs(l log.Logger) {
	libp2pLog.SetDebugLogging()
	libp2pLog.SetPrimaryCore(&zapCore{log: l})
}

type zapCore struct {
	log log.Logger
}

func (z *zapCore) Enabled(level zapcore.Level) bool {
	return true
}

func (z *zapCore) With(zapFields []zapcore.Field) zapcore.Core {
	fields := make(log.Fields, len(zapFields))
	for _, f := range zapFields {
		switch f.Type {
		case zapcore.ArrayMarshalerType:
			// Skip.
		case zapcore.ObjectMarshalerType:
			// Skip.
		case zapcore.InlineMarshalerType:
			// Skip.
		case zapcore.BinaryType:
			fields[f.Key] = f.Interface.([]byte)
		case zapcore.BoolType:
			fields[f.Key] = f.Integer == 1
		case zapcore.ByteStringType:
			fields[f.Key] = f.Interface.([]byte)
		case zapcore.Complex128Type:
			fields[f.Key] = f.Interface.(complex128)
		case zapcore.Complex64Type:
			fields[f.Key] = f.Interface.(complex64)
		case zapcore.DurationType:
			fields[f.Key] = time.Duration(f.Integer)
		case zapcore.Float64Type:
			fields[f.Key] = math.Float64frombits(uint64(f.Integer))
		case zapcore.Float32Type:
			fields[f.Key] = math.Float32frombits(uint32(f.Integer))
		case zapcore.Int64Type:
			fields[f.Key] = f.Integer
		case zapcore.Int32Type:
			fields[f.Key] = int32(f.Integer)
		case zapcore.Int16Type:
			fields[f.Key] = int16(f.Integer)
		case zapcore.Int8Type:
			fields[f.Key] = int8(f.Integer)
		case zapcore.StringType:
			fields[f.Key] = f.String
			z.log.WithField(f.Key, f.String)
		case zapcore.TimeType:
			if f.Interface != nil {
				fields[f.Key] = time.Unix(0, f.Integer).In(f.Interface.(*time.Location))
			} else {
				fields[f.Key] = time.Unix(0, f.Integer)
			}
		case zapcore.TimeFullType:
			fields[f.Key] = f.Interface.(time.Time)
		case zapcore.Uint64Type:
			fields[f.Key] = uint64(f.Integer)
		case zapcore.Uint32Type:
			fields[f.Key] = uint32(f.Integer)
		case zapcore.Uint16Type:
			fields[f.Key] = uint16(f.Integer)
		case zapcore.Uint8Type:
			fields[f.Key] = uint8(f.Integer)
		case zapcore.UintptrType:
			// Skip.
		case zapcore.ReflectType:
			// Skip.
		case zapcore.NamespaceType:
			// Skip.
		case zapcore.StringerType:
			fields[f.Key] = f.Interface.(fmt.Stringer).String()
		case zapcore.ErrorType:
			fields[f.Key] = f.Interface.(error).Error()
		case zapcore.SkipType:
			// Skip.
		}
	}
	return &zapCore{log: z.log.WithFields(fields)}
}

func (z *zapCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(ent, z)
}

func (z *zapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	z.With(fields).(*zapCore).log.
		WithField("TAG", "LIBP2P").
		Debug(fmt.Sprintf("Internal: %s", entry.Message))
	return nil
}

func (z *zapCore) Sync() error {
	return nil
}
