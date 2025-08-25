package rpc

var Packers = map[string]Packer{}

func RegisterPacker(packer Packer) {
	Packers[packer.Name()] = packer
}
func SelectPacker(fun *FunInfo) Packer {
	// 选择逻辑
	for _, packer := range Packers {
		if packer.Match(fun) {
			return packer
		}
	}
	return nil
}
