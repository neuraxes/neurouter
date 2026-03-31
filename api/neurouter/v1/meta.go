package v1

func (x *Content) Meta(key string) string {
	if x == nil {
		return ""
	}
	return x.Metadata[key]
}

func (x *Content) SetMeta(key, value string) {
	if x == nil {
		return
	}
	if x.Metadata == nil {
		x.Metadata = make(map[string]string)
	}
	x.Metadata[key] = value
}

func (x *Message) Meta(key string) string {
	if x == nil {
		return ""
	}
	return x.Metadata[key]
}

func (x *Message) SetMeta(key, value string) {
	if x == nil {
		return
	}
	if x.Metadata == nil {
		x.Metadata = make(map[string]string)
	}
	x.Metadata[key] = value
}

func (x *ChatReq) Meta(key string) string {
	if x == nil {
		return ""
	}
	return x.Metadata[key]
}

func (x *ChatReq) SetMeta(key, value string) {
	if x == nil {
		return
	}
	if x.Metadata == nil {
		x.Metadata = make(map[string]string)
	}
	x.Metadata[key] = value
}
