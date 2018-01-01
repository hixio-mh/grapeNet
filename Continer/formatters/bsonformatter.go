package formatters

/// 使用Bson来打包整个Item
/// 数据会被序列化为Bson然后保存

import (
	"fmt"

	util "github.com/koangel/grapeNet/Utils"
	"gopkg.in/mgo.v2/bson"
)

type BsonFormatter struct{}

func (b *BsonFormatter) To(val, info interface{}) (out []byte, err error) {
	err = nil
	out = []byte{}

	if val == nil {
		err = fmt.Errorf("value is nil...")
		return
	}

	arr := [][]byte{}
	vb, berr := bson.Marshal(val)
	if err != nil {
		err = berr
		return
	}

	arr = append(arr, vb)

	if info != nil {
		ib, ierr := bson.Marshal(info)
		if ierr != nil {
			err = ierr
			return
		}

		arr = append(arr, ib)
	}

	out = util.MergeBinary(arr...)
	err = nil
	return
}

func (b *BsonFormatter) From(src []byte, val, info interface{}) error {
	out := util.SplitBinary(src)
	if len(out) == 0 {
		return fmt.Errorf("source data error,parse faild...")
	}

	if val == nil {
		return fmt.Errorf("value is nil...")
	}

	err := bson.Unmarshal(out[0], val)
	if err != nil {
		return err
	}

	if len(out) > 1 {
		err = bson.Unmarshal(out[1], info)
		if err != nil {
			return err
		}
	}

	return nil
}
