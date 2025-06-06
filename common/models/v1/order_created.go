// Code generated by github.com/actgardner/gogen-avro/v7. DO NOT EDIT.
/*
 * SOURCE:
 *     order_created_v1.avsc
 */
package models_v1

import (
	"github.com/actgardner/gogen-avro/v7/compiler"
	"github.com/actgardner/gogen-avro/v7/vm"
	"github.com/actgardner/gogen-avro/v7/vm/types"
	"io"
)

type OrderCreated struct {
	OrderID string `json:"orderID"`

	UserID string `json:"userID"`

	Items []string `json:"items"`

	Total float64 `json:"total"`
}

const OrderCreatedAvroCRC64Fingerprint = "dN\b\xe7\xabo\xe3\xaf"

func NewOrderCreated() *OrderCreated {
	return &OrderCreated{}
}

func DeserializeOrderCreated(r io.Reader) (*OrderCreated, error) {
	t := NewOrderCreated()
	deser, err := compiler.CompileSchemaBytes([]byte(t.Schema()), []byte(t.Schema()))
	if err != nil {
		return nil, err
	}

	err = vm.Eval(r, deser, t)
	if err != nil {
		return nil, err
	}
	return t, err
}

func DeserializeOrderCreatedFromSchema(r io.Reader, schema string) (*OrderCreated, error) {
	t := NewOrderCreated()

	deser, err := compiler.CompileSchemaBytes([]byte(schema), []byte(t.Schema()))
	if err != nil {
		return nil, err
	}

	err = vm.Eval(r, deser, t)
	if err != nil {
		return nil, err
	}
	return t, err
}

func writeOrderCreated(r *OrderCreated, w io.Writer) error {
	var err error
	err = vm.WriteString(r.OrderID, w)
	if err != nil {
		return err
	}
	err = vm.WriteString(r.UserID, w)
	if err != nil {
		return err
	}
	err = writeArrayString(r.Items, w)
	if err != nil {
		return err
	}
	err = vm.WriteDouble(r.Total, w)
	if err != nil {
		return err
	}
	return err
}

func (r *OrderCreated) Serialize(w io.Writer) error {
	return writeOrderCreated(r, w)
}

func (r *OrderCreated) Schema() string {
	return "{\"fields\":[{\"name\":\"orderID\",\"type\":\"string\"},{\"name\":\"userID\",\"type\":\"string\"},{\"name\":\"items\",\"type\":{\"items\":\"string\",\"type\":\"array\"}},{\"name\":\"total\",\"type\":\"double\"}],\"name\":\"ecommerce.OrderCreated\",\"type\":\"record\"}"
}

func (r *OrderCreated) SchemaName() string {
	return "ecommerce.OrderCreated"
}

func (_ *OrderCreated) SetBoolean(v bool)    { panic("Unsupported operation") }
func (_ *OrderCreated) SetInt(v int32)       { panic("Unsupported operation") }
func (_ *OrderCreated) SetLong(v int64)      { panic("Unsupported operation") }
func (_ *OrderCreated) SetFloat(v float32)   { panic("Unsupported operation") }
func (_ *OrderCreated) SetDouble(v float64)  { panic("Unsupported operation") }
func (_ *OrderCreated) SetBytes(v []byte)    { panic("Unsupported operation") }
func (_ *OrderCreated) SetString(v string)   { panic("Unsupported operation") }
func (_ *OrderCreated) SetUnionElem(v int64) { panic("Unsupported operation") }

func (r *OrderCreated) Get(i int) types.Field {
	switch i {
	case 0:
		return &types.String{Target: &r.OrderID}
	case 1:
		return &types.String{Target: &r.UserID}
	case 2:
		r.Items = make([]string, 0)

		return &ArrayStringWrapper{Target: &r.Items}
	case 3:
		return &types.Double{Target: &r.Total}
	}
	panic("Unknown field index")
}

func (r *OrderCreated) SetDefault(i int) {
	switch i {
	}
	panic("Unknown field index")
}

func (r *OrderCreated) NullField(i int) {
	switch i {
	}
	panic("Not a nullable field index")
}

func (_ *OrderCreated) AppendMap(key string) types.Field { panic("Unsupported operation") }
func (_ *OrderCreated) AppendArray() types.Field         { panic("Unsupported operation") }
func (_ *OrderCreated) Finalize()                        {}

func (_ *OrderCreated) AvroCRC64Fingerprint() []byte {
	return []byte(OrderCreatedAvroCRC64Fingerprint)
}
