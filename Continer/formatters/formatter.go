package formatters

/// 编写标准，请务必再返回前使用,utils库来合并所有的[]byte，否则可能导致无法分析。
/// 具体请参考BsonFormatter

type ItemFormatter interface {
	To(val, info interface{}) (out []byte, err error)
	From(src []byte, val, info interface{}) error
}
