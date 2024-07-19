package replayor

type TxTrace struct {
	Gas        uint64 `json:"gas"`
	Failed     bool   `json:"failed"`
	StructLogs []struct {
		Op      string `json:"op"`
		Gas     uint64 `json:"gas"`
		GasCost uint64 `json:"gasCost"`
		Refund  uint64 `json:"refund"`
		Depth   uint64 `json:"depth"`
	} `json:"structLogs"`
}
