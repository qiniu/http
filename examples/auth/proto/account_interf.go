package proto

// --------------------------------------------------------------------

type UserInfo struct {
	Uid    uint32 `json:"uid"`
	Utype  uint32 `json:"ut"`
	Appid  uint64 `json:"app,omitempty"`
	Access string `json:"ak,omitempty"`

	EndUser string `json:"eu,omitempty"`
}

type SudoerInfo struct {
	UserInfo
	Sudoer  uint32 `json:"suid,omitempty"`
	UtypeSu uint32 `json:"sut,omitempty"`
}

// --------------------------------------------------------------------
