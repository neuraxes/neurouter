package v1

// IsReasoning reports whether the content belongs to the reasoning phase.
func (x *Content) IsReasoning() bool {
	return x.GetPhase() == ContentPhase_CONTENT_PHASE_REASONING
}

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
