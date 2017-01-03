package rorm

import (
	"testing"
	"time"

	"encoding/json"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStringQuerySet(t *testing.T) {
	rormer := NewROrm()
	qs := rormer.QueryString("Hello")
	qs = qs.Protect(180 * time.Second)

	testModel := TestModel{
		Name: "Youhuhu",
		Age:  18,
	}

	err := qs.Set(testModel, 180*time.Second)

	Convey("Test Set:", t, func() {
		So(err, ShouldBeNil)

		var scn TestModel
		qs.Scan(&scn)
		Convey("Test Scan: ", func() {
			So(scn.Name, ShouldEqual, testModel.Name)
			So(scn.Age, ShouldEqual, testModel.Age)
		})
	})

}

type TestModel struct {
	Name string
	Age  int
}

func (this *TestModel) MarshalBinary() (data []byte, err error) {
	return json.Marshal(this)
}

func (this *TestModel) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, this)
}
