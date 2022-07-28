package expand

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/DataHighway-DHX/bifrost-go/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/huandu/xstrings"
)

/*
对metadata进行扩展，添加一些实用的功能
由于大多数的波卡链都升级到了v11和v12，所以只对大于v11的链处理
*/
type MetadataExpand struct {
	meta *types.Metadata
	MV   iMetaVersion
}
type iMetaVersion interface {
	GetCallIndex(moduleName, fn string) (callIdx string, err error)
	FindNameByCallIndex(callIdx string) (moduleName, fn string, err error)
	GetConstants(modName, constantsName string) (constantsType string, constantsValue []byte, err error)
}

func NewMetadataExpand(meta *types.Metadata) (*MetadataExpand, error) {
	me := new(MetadataExpand)
	me.meta = meta
	switch meta.Version {
	case 11:
		me.MV = newV11(meta.AsMetadataV11.Modules)
	case 12:
		me.MV = newV12(meta.AsMetadataV12.Modules)
	case 13:
		me.MV = newV13(meta.AsMetadataV13.Modules)
	case 14:
		me.MV = newV14(meta)
	default:
		return nil, errors.New("metadata version is not v11 or v12 or 13 or 14")
	}

	return me, nil
}

/*
func: 没办法，只能适应之前写的啰
author: flynn
date: 2021/10/29
*/
// type v14Upgrade struct {
// 	meta *types.Metadata
// }

// func newV14UpGrade(meta *types.Metadata) *v14Upgrade {
// 	v := new(v14Upgrade)
// 	v.meta = meta
// 	return v
// }

// // TODO:
// func (v v14Upgrade) GetCallIndex(moduleName, fn string) (callIdx string, err error) {
// 	return v.meta.GetCallIndex(moduleName, fn)
// }

// func (v v14Upgrade) FindNameByCallIndex(callIdx string) (moduleName, fn string, err error) {
// 	for _, mod := range m.Pallets {
// 		if !mod.HasEvents {
// 			continue
// 		}
// 		if mod.Index != NewU8(eventID[0]) {
// 			continue
// 		}
// 		eventType := mod.Events.Type.Int64()

// 		if typ, ok := m.EfficientLookup[eventType]; ok {
// 			if len(typ.Def.Variant.Variants) > 0 {
// 				for _, vars := range typ.Def.Variant.Variants {
// 					if uint8(vars.Index) == eventID[1] {
// 						return mod.Name, vars.Name, nil
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return "", "", fmt.Errorf("module index %v out of range", eventID[0])

// 	return "", "", fmt.Errorf("do not find this callInx info: %s", callIdx)
// }

// func (v v14Upgrade) GetConstants(modName, constantsName string) (constantsType string, constantsValue []byte, err error) {
// 	v.meta.FindConstantValue(modName, constantsName)
// 	v.meta.AsMetadataV14.Type
// 	return v.meta.GetConstants(modName, constantsName)
// }

func newV14(module *types.Metadata) *v14 {
	v := new(v14)
	v.module = *module
	return v
}

type v14 struct {
	module types.Metadata
}

func (v v14) GetCallIndex(moduleName, fn string) (callIdx string, err error) {
	fmt.Println("------- GetCallIndex")
	defer func() {
		if errs := recover(); errs != nil {
			callIdx = ""
			err = fmt.Errorf("catch panic ,err=%v", errs)
		}
	}()
	ci, err := v.module.FindCallIndex(fmt.Sprintf("%s.%s", moduleName, fn))
	fmt.Printf("%v ------ %v------- ---- FindCallIndex", ci, err)
	return xstrings.RightJustify(fmt.Sprintf("%x", ci.SectionIndex), 2, "0") + xstrings.RightJustify(fmt.Sprintf("%x", ci.MethodIndex), 2, "0"), err
}

func (v v14) FindNameByCallIndex(callIdx string) (moduleName, fn string, err error) {
	if len(callIdx) != 4 {
		return "", "", fmt.Errorf("call index length is not equal 4: length: %d", len(callIdx))
	}
	data, err := hex.DecodeString(callIdx)
	if err != nil {
		return "", "", fmt.Errorf("call index is not hex string")
	}

	mn, fun, err := v.module.FindEventNamesForEventID(types.EventID{
		data[0], data[1],
	})

	fmt.Printf("%s ------ %s------- %s ---- FindNameByCallIndex", string(mn), string(fun), err)

	return string(mn), string(fun), err
}

func (v v14) GetConstants(modName, constantsName string) (constantsType string, constantsValue []byte, err error) {
	// defer func() {
	// 	if errs := recover(); errs != nil {
	// 		err = fmt.Errorf("catch panic ,err=%v", errs)
	// 	}
	// }()

	fmt.Println("------- GetConstants")
	val, err := v.module.FindConstantValue(modName, constantsName)

	fmt.Printf("%s ------ %s------- ---- FindConstantValue", val, err)

	return string(rune(v.module.AsMetadataV14.Type.Int64())), val, err
}

type v11 struct {
	module []types.ModuleMetadataV10
}

func (v v11) GetCallIndex(moduleName, fn string) (callIdx string, err error) {
	//避免指针为空
	defer func() {
		if errs := recover(); errs != nil {
			callIdx = ""
			err = fmt.Errorf("catch panic ,err=%v", errs)
		}
	}()
	mi := uint8(0)
	for _, mod := range v.module {
		if !mod.HasCalls {
			continue
		}
		if string(mod.Name) != moduleName {
			mi++
			continue
		}
		for ci, f := range mod.Calls {
			if string(f.Name) == fn {
				return xstrings.RightJustify(utils.IntToHex(mi), 2, "0") + xstrings.RightJustify(utils.IntToHex(ci), 2, "0"), nil
			}
		}
	}
	return "", fmt.Errorf("do not find this call index")
}

func (v v11) FindNameByCallIndex(callIdx string) (moduleName, fn string, err error) {
	if len(callIdx) != 4 {
		return "", "", fmt.Errorf("call index length is not equal 4: length: %d", len(callIdx))
	}
	data, err := hex.DecodeString(callIdx)
	if err != nil {
		return "", "", fmt.Errorf("call index is not hex string")
	}
	mi := int(data[0])
	ci := int(data[1])
	for i, mod := range v.module {
		if !mod.HasCalls {
			continue
		}
		if i == int(mi) {

			for j, call := range mod.Calls {
				if j == int(ci) {
					moduleName = string(mod.Name)
					fn = string(call.Name)
					return
				}
			}
		}
	}
	return "", "", fmt.Errorf("do not find this callInx info: %s", callIdx)
}

func (v v11) GetConstants(modName, constantsName string) (constantsType string, constantsValue []byte, err error) {
	defer func() {
		if errs := recover(); errs != nil {
			err = fmt.Errorf("catch panic ,err=%v", errs)
		}
	}()
	for _, mod := range v.module {
		if modName == string(mod.Name) {
			for _, constants := range mod.Constants {
				if string(constants.Name) == constantsName {
					constantsType = string(constants.Type)
					constantsValue = constants.Value
					return constantsType, constantsValue, nil
				}
			}
		}
	}
	return "", nil, fmt.Errorf("do not find this constants,moduleName=%s,"+
		"constantsName=%s", modName, constantsName)
}

func newV11(module []types.ModuleMetadataV10) *v11 {
	v := new(v11)
	v.module = module
	return v
}

type v12 struct {
	module []types.ModuleMetadataV12
}

func (v v12) FindNameByCallIndex(callIdx string) (moduleName, fn string, err error) {
	if len(callIdx) != 4 {
		return "", "", fmt.Errorf("call index length is not equal 4: length: %d", len(callIdx))
	}
	data, err := hex.DecodeString(callIdx)
	if err != nil {
		return "", "", fmt.Errorf("call index is not hex string")
	}
	for _, mod := range v.module {
		if !mod.HasCalls {
			continue
		}
		if mod.Index == data[0] {

			for j, call := range mod.Calls {
				if j == int(data[1]) {
					moduleName = string(mod.Name)
					fn = string(call.Name)
					return
				}
			}
		}
	}
	return "", "", fmt.Errorf("do not find this callInx info: %s", callIdx)
}

func (v v12) GetCallIndex(moduleName, fn string) (callIdx string, err error) {
	//避免指针为空
	defer func() {
		if errs := recover(); errs != nil {
			callIdx = ""
			err = fmt.Errorf("catch panic ,err=%v", errs)
		}
	}()
	for _, mod := range v.module {
		if !mod.HasCalls {
			continue
		}
		if string(mod.Name) != moduleName {

			continue
		}
		for ci, f := range mod.Calls {
			if string(f.Name) == fn {
				return xstrings.RightJustify(utils.IntToHex(mod.Index), 2, "0") + xstrings.RightJustify(utils.IntToHex(ci), 2, "0"), nil
			}
		}
	}
	return "", fmt.Errorf("do not find this call index")
}
func (v v12) GetConstants(modName, constantsName string) (constantsType string, constantsValue []byte, err error) {
	defer func() {
		if errs := recover(); errs != nil {
			err = fmt.Errorf("catch panic ,err=%v", errs)
		}
	}()
	for _, mod := range v.module {
		if modName == string(mod.Name) {
			for _, constants := range mod.Constants {
				if string(constants.Name) == constantsName {
					constantsType = string(constants.Type)
					constantsValue = constants.Value
					return constantsType, constantsValue, nil
				}
			}
		}
	}
	return "", nil, fmt.Errorf("do not find this constants,moduleName=%s,"+
		"constantsName=%s", modName, constantsName)
}

func newV12(module []types.ModuleMetadataV12) *v12 {
	v := new(v12)
	v.module = module
	return v
}

type v13 struct {
	module []types.ModuleMetadataV13
}

func newV13(module []types.ModuleMetadataV13) *v13 {
	v := new(v13)
	v.module = module
	return v
}
func (v v13) FindNameByCallIndex(callIdx string) (moduleName, fn string, err error) {
	if len(callIdx) != 4 {
		return "", "", fmt.Errorf("call index length is not equal 4: length: %d", len(callIdx))
	}
	data, err := hex.DecodeString(callIdx)
	if err != nil {
		return "", "", fmt.Errorf("call index is not hex string")
	}
	for _, mod := range v.module {
		if !mod.HasCalls {
			continue
		}
		if mod.Index == data[0] {

			for j, call := range mod.Calls {
				if j == int(data[1]) {
					moduleName = string(mod.Name)
					fn = string(call.Name)
					return
				}
			}
		}
	}
	return "", "", fmt.Errorf("do not find this callInx info: %s", callIdx)
}

func (v v13) GetCallIndex(moduleName, fn string) (callIdx string, err error) {
	//避免指针为空
	defer func() {
		if errs := recover(); errs != nil {
			callIdx = ""
			err = fmt.Errorf("catch panic ,err=%v", errs)
		}
	}()
	for _, mod := range v.module {
		if !mod.HasCalls {
			continue
		}
		if string(mod.Name) != moduleName {

			continue
		}
		for ci, f := range mod.Calls {
			if string(f.Name) == fn {
				return xstrings.RightJustify(utils.IntToHex(mod.Index), 2, "0") + xstrings.RightJustify(utils.IntToHex(ci), 2, "0"), nil
			}
		}
	}
	return "", fmt.Errorf("do not find this call index")
}
func (v v13) GetConstants(modName, constantsName string) (constantsType string, constantsValue []byte, err error) {
	defer func() {
		if errs := recover(); errs != nil {
			err = fmt.Errorf("catch panic ,err=%v", errs)
		}
	}()
	for _, mod := range v.module {
		if modName == string(mod.Name) {
			for _, constants := range mod.Constants {
				if string(constants.Name) == constantsName {
					constantsType = string(constants.Type)
					constantsValue = constants.Value
					return constantsType, constantsValue, nil
				}
			}
		}
	}
	return "", nil, fmt.Errorf("do not find this constants,moduleName=%s,"+
		"constantsName=%s", modName, constantsName)
}

/*
Balances.transfer
*/
func (e *MetadataExpand) BalanceTransferCall(to string, amount uint64) (types.Call, error) {
	var (
		call types.Call
	)
	callIdx, err := e.MV.GetCallIndex("Balances", "transfer")
	if err != nil {
		return call, err
	}
	recipientPubkey := utils.AddressToPublicKey(to)
	var ma MultiAddress
	ma.SetTypes(0)
	ma.AccountId = types.NewAccountID(types.MustHexDecodeString(recipientPubkey))
	return NewCall(callIdx, ma,
		types.NewUCompactFromUInt(amount))

}

/*
Balances.transfer_keep_alive
*/
func (e *MetadataExpand) BalanceTransferKeepAliveCall(to string, amount uint64) (types.Call, error) {
	var (
		call types.Call
	)
	callIdx, err := e.MV.GetCallIndex("Balances", "transfer_keep_alive")
	if err != nil {
		return call, err
	}
	recipientPubkey := utils.AddressToPublicKey(to)
	var ma MultiAddress
	ma.SetTypes(0)
	ma.AccountId = types.NewAccountID(types.MustHexDecodeString(recipientPubkey))
	return NewCall(callIdx, ma,
		types.NewUCompactFromUInt(amount))

}

/*
Utility.batch
keepAlive: true->Balances.transfer_keep_alive	false->Balances.transfer
*/
func (e *MetadataExpand) UtilityBatchTxCall(toAmount map[string]uint64, keepAlive bool) (types.Call, error) {
	var (
		call types.Call
		err  error
	)
	if len(toAmount) == 0 {
		return call, errors.New("toAmount is null")
	}
	var calls []types.Call
	for to, amount := range toAmount {
		var (
			btCall types.Call
		)
		if keepAlive {
			btCall, err = e.BalanceTransferKeepAliveCall(to, amount)
		} else {
			btCall, err = e.BalanceTransferCall(to, amount)
		}
		if err != nil {
			return call, err
		}
		calls = append(calls, btCall)
	}
	callIdx, err := e.MV.GetCallIndex("Utility", "batch")
	if err != nil {
		return call, err
	}
	return NewCall(callIdx, calls)
}

/*
transfer with memo
*/
func (e *MetadataExpand) UtilityBatchTxWithMemo(to, memo string, amount uint64) (types.Call, error) {
	var (
		call types.Call
	)
	btCall, err := e.BalanceTransferCall(to, amount)
	if err != nil {
		return call, err
	}
	smCallIdx, err := e.MV.GetCallIndex("System", "remark")
	if err != nil {
		return call, err
	}
	smCall, err := NewCall(smCallIdx, memo)
	ubCallIdx, err := e.MV.GetCallIndex("Utility", "batch")
	if err != nil {
		return call, err
	}
	return NewCall(ubCallIdx, btCall, smCall)
}
