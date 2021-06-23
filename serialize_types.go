package types

import (
	"bytes"
	"errors"
	"github.com/nervosnetwork/io"
)

func (h Hash) Serialize() ([]byte, error) {
	return h.Bytes(), nil
}

func (h *Hash) DeSerialize(br *io.BinaryReader)  {
	bytes := make([]byte, HashLength)
	br.ReadLE(bytes)
	h.SetBytes(bytes)
}

func (t ScriptHashType) Serialize() ([]byte, error) {
	if t == HashTypeData {
		return []byte{00}, nil
	} else if t == HashTypeType {
		return []byte{01}, nil
	}
	return nil, errors.New("invalid script hash type")
}

func (t *ScriptHashType) DeSerialize(br *io.BinaryReader)  {
	typeDep := uint8(0)

	br.ReadLE(&typeDep)
	if typeDep == 0 {
		*t = HashTypeData
	}else{
		*t = HashTypeType
	}
}

// Serialize dep type
func (t DepType) Serialize() ([]byte, error) {
	if t == DepTypeCode {
		return []byte{00}, nil
	} else if t == DepTypeDepGroup {
		return []byte{01}, nil
	}
	return nil, errors.New("invalid dep group")
}

func (t *DepType) DeSerialize(br *io.BinaryReader)  {
	typeDep := uint8(0)

	br.ReadLE(&typeDep)
	if typeDep == 0 {
		*t = DepTypeCode
	}else{
		*t = DepTypeDepGroup
	}

}

// Serialize script
func (script *Script) Serialize() ([]byte, error) {
	h, err := script.CodeHash.Serialize()
	if err != nil {
		return nil, err
	}
	t, err := script.HashType.Serialize()
	if err != nil {
		return nil, err
	}
	a := SerializeBytes(script.Args)
	return SerializeTable([][]byte{h, t, a}), nil
}

func (script *Script) DeSerialize(br *io.BinaryReader) {

	vBytes := DeSerializeDynVec(br)

	for nIndex :=0; nIndex< int(len(vBytes)); nIndex++{

		oneByte := vBytes[nIndex]
		br2 := io.NewBinaryReaderFromBuf(oneByte)
		if nIndex == 0 {
			script.CodeHash.DeSerialize(br2)
		}else if nIndex == 1 {
			script.HashType.DeSerialize(br2)
		}else if nIndex == 2 {
			byteLen := uint32(0)
			br2.ReadLE(&byteLen)
			argByte := make([]byte, byteLen)
			br2.ReadLE(argByte)
			script.Args = argByte
		}
	}

}

// Serialize outpoint
func (o *OutPoint) Serialize() ([]byte, error) {
	h, err := o.TxHash.Serialize()
	if err != nil {
		return nil, err
	}

	i := SerializeUint(o.Index)
	b := new(bytes.Buffer)

	b.Write(h)
	b.Write(i)

	return b.Bytes(), nil
}

func (o *OutPoint) DeSerialize (br *io.BinaryReader)  {

	vByte := make([]byte, HashLength)
	br.ReadLE(vByte)
	o.TxHash.SetBytes(vByte)
	nIn := uint32(0)
	br.ReadLE(&nIn)
	o.Index = uint(nIn)
}

// Serialize cell input
func (i *CellInput) Serialize() ([]byte, error) {
	s := SerializeUint64(i.Since)

	o, err := i.PreviousOutput.Serialize()
	if err != nil {
		return nil, err
	}

	return SerializeStruct([][]byte{s, o}), nil
}

func (i *CellInput) DeSerialize(br *io.BinaryReader)  {
	op := OutPoint{}
	i.PreviousOutput = &op

	br.ReadLE(&i.Since)
	i.PreviousOutput.DeSerialize(br)
}

// Serialize cell output
func (o *CellOutput) Serialize() ([]byte, error) {
	c := SerializeUint64(o.Capacity)

	l, err := o.Lock.Serialize()
	if err != nil {
		return nil, err
	}

	t, err := SerializeOption(o.Type)
	if err != nil {
		return nil, err
	}

	return SerializeTable([][]byte{c, l, t}), nil
}
func (o *CellOutput) DeSerialize(br * io.BinaryReader)  {

	lock := Script{}
	o.Lock = &lock

    vBytes := DeSerializeDynVec(br)
	for nIndex := 0; nIndex< int(len(vBytes)); nIndex++{

		oneByte := vBytes[nIndex]
		br2 := io.NewBinaryReaderFromBuf(oneByte)
		if nIndex == 0 {
			br2.ReadLE(&o.Capacity)
		}else if nIndex == 1 {
			o.Lock.DeSerialize(br2)
		}else if nIndex == 2 {
			if len(oneByte) > 0{
				typet := Script{}
				o.Type = &typet
				o.Type.DeSerialize(br2)
			}

		}
	}

}

// Serialize cell dep
func (d *CellDep) Serialize() ([]byte, error) {
	o, err := d.OutPoint.Serialize()
	if err != nil {
		return nil, err
	}

	dd, err := d.DepType.Serialize()
	if err != nil {
		return nil, err
	}

	return SerializeStruct([][]byte{o, dd}), nil
}

func (d *CellDep) DeSerialize(br *io.BinaryReader)  {
	op := OutPoint{}
	d.OutPoint = &op
	d.OutPoint.DeSerialize(br)
	d.DepType.DeSerialize(br)
}

// Serialize transaction
func (t *Transaction) Serialize() ([]byte, error) {
	v := SerializeUint(t.Version)

	// Ok, no way around this
	deps := make([]Serializer, len(t.CellDeps))
	for i, v := range t.CellDeps {
		deps[i] = v
	}
	cds, err := SerializeArray(deps)
	if err != nil {
		return nil, err
	}
	cdsBytes := SerializeFixVec(cds)

	hds := make([][]byte, len(t.HeaderDeps))
	for i := 0; i < len(t.HeaderDeps); i++ {
		hd, err := t.HeaderDeps[i].Serialize()
		if err != nil {
			return nil, err
		}

		hds[i] = hd
	}
	hdsBytes := SerializeFixVec(hds)

	ips := make([][]byte, len(t.Inputs))
	for i := 0; i < len(t.Inputs); i++ {
		ip, err := t.Inputs[i].Serialize()
		if err != nil {
			return nil, err
		}

		ips[i] = ip
	}
	ipsBytes := SerializeFixVec(ips)

	ops := make([][]byte, len(t.Outputs))
	for i := 0; i < len(t.Outputs); i++ {
		op, err := t.Outputs[i].Serialize()
		if err != nil {
			return nil, err
		}

		ops[i] = op
	}
	opsBytes := SerializeDynVec(ops)

	ods := make([][]byte, len(t.OutputsData))
	for i := 0; i < len(t.OutputsData); i++ {
		od := SerializeBytes(t.OutputsData[i])

		ods[i] = od
	}
	odsBytes := SerializeDynVec(ods)

	fields := [][]byte{v, cdsBytes, hdsBytes, ipsBytes, opsBytes, odsBytes}
	return SerializeTable(fields), nil
}

func (t *Transaction) SerializeFull() ([]byte, error) {
	v := SerializeUint(t.Version)

	// Ok, no way around this
	deps := make([]Serializer, len(t.CellDeps))
	for i, v := range t.CellDeps {
		deps[i] = v
	}
	cds, err := SerializeArray(deps)
	if err != nil {
		return nil, err
	}
	cdsBytes := SerializeFixVec(cds)

	hds := make([][]byte, len(t.HeaderDeps))
	for i := 0; i < len(t.HeaderDeps); i++ {
		hd, err := t.HeaderDeps[i].Serialize()
		if err != nil {
			return nil, err
		}

		hds[i] = hd
	}
	hdsBytes := SerializeFixVec(hds)

	ips := make([][]byte, len(t.Inputs))
	for i := 0; i < len(t.Inputs); i++ {
		ip, err := t.Inputs[i].Serialize()
		if err != nil {
			return nil, err
		}

		ips[i] = ip
	}
	ipsBytes := SerializeFixVec(ips)

	ops := make([][]byte, len(t.Outputs))
	for i := 0; i < len(t.Outputs); i++ {
		op, err := t.Outputs[i].Serialize()
		if err != nil {
			return nil, err
		}

		ops[i] = op
	}
	opsBytes := SerializeDynVec(ops)

	//
	ods := make([][]byte, len(t.OutputsData))
	for i := 0; i < len(t.OutputsData); i++ {
		od := SerializeBytes(t.OutputsData[i])

		ods[i] = od
	}
	odsBytes := SerializeDynVec(ods)

	//Witnesses
	witness := make([][]byte, len(t.Witnesses))
	for i := 0; i < len(t.Witnesses); i++ {
		od := SerializeBytes(t.Witnesses[i])

		witness[i] = od
	}
	witnessBytes := SerializeDynVec(witness)

	fields := [][]byte{v, cdsBytes, hdsBytes, ipsBytes, opsBytes, odsBytes, witnessBytes}
	return SerializeTable(fields), nil

}

func (w *WitnessArgs) Serialize() ([]byte, error) {
	l, err := SerializeOptionBytes(w.Lock)
	if err != nil {
		return nil, err
	}

	i, err := SerializeOptionBytes(w.InputType)
	if err != nil {
		return nil, err
	}

	o, err := SerializeOptionBytes(w.OutputType)
	if err != nil {
		return nil, err
	}
	return SerializeTable([][]byte{l, i, o}), nil
}

func (t *Transaction) DeserializeFull(b []byte) {
	br := io.NewBinaryReaderFromBuf(b)
	vBytes := DeSerializeDynVec(br)
	t.DeserializeVersion(vBytes[0])
	t.DeserializeCellDep(vBytes[1])
	t.DeserializeHeaderDeps(vBytes[2])
	t.DeserializeInput(vBytes[3])
	t.DeserializeOutput(vBytes[4])
	t.DeserializeOutputData(vBytes[5])
	t.DeserializeWitness(vBytes[6])
}
func (t *Transaction) DeserializeWitness(vbyte []byte){

	if len(vbyte) == 4 {
		return
	}
	br := io.NewBinaryReaderFromBuf(vbyte)
	vBytes := DeSerializeDynVec(br)

	t.Witnesses = make([][]byte,len(vBytes))
	for nIndex:=0; nIndex < int(len(vBytes));nIndex++{

		br1 := io.NewBinaryReaderFromBuf(vBytes[nIndex])
		unlen := uint32(0)
		br1.ReadLE(&unlen)
		bytes := make([]byte, unlen)
		br1.ReadLE(bytes)
		t.Witnesses[nIndex] = bytes
	}
}

func (t *Transaction) DeserializeOutputData(vbyte []byte){

	if len(vbyte) == 4 {
		return
	}
	br := io.NewBinaryReaderFromBuf(vbyte)
	vBytes := DeSerializeDynVec(br)

	t.OutputsData = make([][]byte,len(vBytes))
	for nIndex:=0; nIndex < int(len(vBytes));nIndex++{
		
		br1 := io.NewBinaryReaderFromBuf(vBytes[nIndex])
		unlen := uint32(0)
		br1.ReadLE(&unlen)
		bytes := make([]byte, unlen)
		br1.ReadLE(bytes)
		t.OutputsData[nIndex] = bytes
	}
}

func (t *Transaction) DeserializeOutput(vbyte []byte){

	br := io.NewBinaryReaderFromBuf(vbyte)
	vBytes := DeSerializeDynVec(br)

	t.Outputs = make( []*CellOutput, len(vBytes))
	for nIndex:=0; nIndex < int(len(vBytes));nIndex++{

		obj := CellOutput{}
		br1 := io.NewBinaryReaderFromBuf(vBytes[nIndex])
		obj.DeSerialize(br1)

		t.Outputs[nIndex] = &obj
	}
}


func (t *Transaction) DeserializeInput(vbyte []byte){

	br := io.NewBinaryReaderFromBuf(vbyte)
	unlen := uint32(0)
	br.ReadLE(&unlen)

	if unlen == 0 {
		return
	}
	t.Inputs = make([]*CellInput, unlen)
	for nIndex:=0;nIndex < int(unlen);nIndex++{

		obj := CellInput{}
		n := obj.GetSize()
		oneByte := make([]byte, n)
		br.ReadLE(oneByte)
		br1 := io.NewBinaryReaderFromBuf(oneByte)
		obj.DeSerialize(br1)
		t.Inputs[nIndex] = &obj
	}

}

func (t *Transaction) DeserializeHeaderDeps(vbyte []byte){

	br := io.NewBinaryReaderFromBuf(vbyte)
	unlen := uint32(0)
	br.ReadLE(&unlen)

	if unlen == 0 {
		return
	}

	for nIndex:=0;nIndex < int(unlen);nIndex++{
		oneByte := make([]byte, 32)
		br1 := io.NewBinaryReaderFromBuf(oneByte)
		br1.ReadLE(oneByte)

		t.HeaderDeps[nIndex] = BytesToHash( oneByte)
	}

}

func (t *Transaction) DeserializeVersion(vByte []byte){
	br := io.NewBinaryReaderFromBuf(vByte)
	br.ReadLE(&t.Version)
}

func (t *Transaction) DeserializeCellDep(vbyte []byte){
	
	br := io.NewBinaryReaderFromBuf(vbyte)
	unlen := uint32(0)
	br.ReadLE(&unlen)
	
	if unlen == 0 {
		return
	}

	for nIndex:=0;nIndex < int(unlen);nIndex++{

		obj := CellDep{}
		n := obj.GetSize()
		oneByte := make([]byte, n)
		br.ReadLE(oneByte)
		br1 := io.NewBinaryReaderFromBuf(oneByte)
		obj.DeSerialize(br1)
		t.CellDeps[nIndex] = &obj
	}
	
}
func (op *OutPoint) GetSize() uint32 {
	unlen := uint32(HashLength)
	unlen += 4
	return unlen
}

func (cd *CellDep) GetSize() uint32 {

	unlen := cd.OutPoint.GetSize()
	unlen += 1
	return unlen
}

func (op *CellInput) GetSize() uint32 {

	unlen := uint32(8)
	unlen += op.PreviousOutput.GetSize()
	return unlen
}

func DeSerializeDynVec(br *io.BinaryReader) [][]byte {

	unSizeAll := uint32(0)
	br.ReadLE(&unSizeAll)

	unOffsetFirst := uint32(0)
	br.ReadLE(&unOffsetFirst)

	unlen := (unOffsetFirst -4)/4

	offsets := make([]uint32, unlen)
	fields := make([][]byte, unlen)

	offsets[0] = unOffsetFirst

	for nIndex :=1; nIndex< int(unlen); nIndex++{
		uintOff := uint32(0);
		br.ReadLE(&uintOff)
		offsets[nIndex] = uintOff
	}

	for nIndex:=0; nIndex < int(unlen); nIndex++{
		n := uint32(0)
		if nIndex < int(unlen-1) {
			n = offsets[nIndex+1] - offsets[nIndex]
		}else{
			n = unSizeAll - offsets[nIndex]
		}

		oneByte := make([]byte, n)
		br.ReadLE(oneByte)
		fields[nIndex] = oneByte
	}
	return fields
}
