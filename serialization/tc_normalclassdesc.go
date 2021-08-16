package serialization

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type TCClassDesc struct {
	ClassName         *TCUtf
	SerialVersionUID  int64
	ClassDescFlags    byte
	Fields            []*TCFieldDesc
	ClassAnnotation   []*TCContent
	SuperClassPointer *TCClassPointer
	Handler uint32
}

func (desc *TCClassDesc) ToBytes() []byte {
	var result = []byte{JAVA_TC_CLASSDESC}
	result = append(result, desc.ClassName.ToBytes()...)
	result = append(result, NumberToBytes(desc.SerialVersionUID)...)
	result = append(result, desc.ClassDescFlags)
	result = append(result, NumberToBytes(uint16(len(desc.Fields)))...)
	for _, field := range desc.Fields {
		result = append(result, field.ToBytes()...)
	}
	for _, content := range desc.ClassAnnotation {
		result = append(result, content.ToBytes()...)
	}
	result = append(result, JAVA_TC_ENDBLOCKDATA)
	result = append(result, desc.SuperClassPointer.ToBytes()...)

	return result
}

func (desc *TCClassDesc) ToString() string {
	var b = newPrinter()

	b.printf("TC_CLASSDESC - %s", Hexify(JAVA_TC_CLASSDESC))
	b.increaseIndent()
	b.print("@ClassName")
	b.increaseIndent()
	b.print(desc.ClassName.ToString())
	b.decreaseIndent()
	b.printf("@SerialVersionUID - %v - %s", desc.SerialVersionUID, Hexify(desc.SerialVersionUID))
	b.printf("@Handler - %v", desc.Handler)
	b.printf("@ClassDescFlags - %s - %s", desc.FlagString(), Hexify(desc.ClassDescFlags))
	b.printf("@FieldCount - %d - %s", len(desc.Fields), Hexify(uint16(len(desc.Fields))))
	b.print("[]Fields")
	b.increaseIndent()
	for index, field := range desc.Fields {
		b.printf("Index %d:", index)
		b.increaseIndent()
		b.print(field.ToString())
		b.decreaseIndent()
	}
	b.decreaseIndent()
	b.print("[]ClassAnnotations")
	b.increaseIndent()
	for index, content := range desc.ClassAnnotation {
		b.printf("Index %d:", index)
		b.increaseIndent()
		b.print(content.ToString())
		b.decreaseIndent()
	}
	b.printf("TC_ENDBLOCKDATA - %s", Hexify(JAVA_TC_ENDBLOCKDATA))
	b.decreaseIndent()
	b.print("@SuperClassDesc")
	b.increaseIndent()
	b.print(desc.SuperClassPointer.ToString())

	return b.String()
}

// HasFlag Check if a TCClassDesc object has a flag
func (desc *TCClassDesc) HasFlag(flag byte) bool {
	return (desc.ClassDescFlags & flag) == flag
}

func (desc *TCClassDesc) FlagString() string {
	var descFlags []string
	if desc.HasFlag(JAVA_SC_SERIALIZABLE) {
		descFlags = append(descFlags, "SC_SERIALIZABLE")
	}
	if desc.HasFlag(JAVA_SC_WRITE_METHOD) {
		descFlags = append(descFlags, "SC_WRITE_METHOD")
	}
	if desc.HasFlag(JAVA_SC_EXTERNALIZABLE) {
		descFlags = append(descFlags, "SC_EXTERNALIZABLE")
	}
	if desc.HasFlag(JAVA_SC_BLOCK_DATA) {
		descFlags = append(descFlags, "SC_BLOCK_DATA")
	}

	return strings.Join(descFlags, "|")
}

func readTCNormalClassDesc(stream *ObjectStream) (*TCClassDesc, error) {
	var err error
	var classDesc = new(TCClassDesc)

	// read JAVA_TC_CLASSDESC flag
	_, _ = stream.ReadN(1)

	// className
	classDesc.ClassName, err = readUTF(stream)
	if err != nil {
		return nil, err
	}

	// serialVersionUID
	classDesc.SerialVersionUID, err = readSerialVersionUID(stream)
	if err != nil {
		return nil, err
	}

	// add handle to reference
	stream.AddReference(classDesc)

	// ------ classDescInfo -------
	// classDescFlags
	bs, err := stream.ReadN(1)
	if err != nil {
		return nil, fmt.Errorf("read JAVA_TC_CLASSDESC failed on index %v", stream.CurrentIndex())
	}
	classDesc.ClassDescFlags = bs[0]

	// fields
	classDesc.Fields, err = readTCFields(stream)
	if err != nil {
		return nil, err
	}

	// classAnnotation
	classDesc.ClassAnnotation, err = readTCAnnotation(stream)
	if err != nil {
		return nil, err
	}

	// superClassDesc
	classDesc.SuperClassPointer, err = readTCClassPointer(stream)
	if err != nil {
		return nil, err
	}

	return classDesc, nil
}

func readTCAnnotation(stream *ObjectStream) ([]*TCContent, error) {
	var contents []*TCContent
	for {
		bs, err := stream.PeekN(1)
		if err != nil {
			return nil, fmt.Errorf("read classAnnotation failed on index %v", stream.CurrentIndex())
		}

		if bs[0] == JAVA_TC_ENDBLOCKDATA {
			_, _ = stream.ReadN(1)
			break
		}

		content, err := readTCContent(stream)
		if err != nil {
			return nil, err
		}

		contents = append(contents, content)
	}

	return contents, nil
}

func readTCFields(stream *ObjectStream) ([]*TCFieldDesc, error) {
	var bs []byte
	var err error
	var fields []*TCFieldDesc

	bs, err = stream.ReadN(2)
	if err != nil {
		return nil, fmt.Errorf("read JAVA_TC_CLASSDESC failed on index %v", stream.CurrentIndex())
	}
	fieldsLength := binary.BigEndian.Uint16(bs)

	for i := uint16(0); i < fieldsLength; i++ {
		field, err := readTCField(stream)
		if err != nil {
			return nil, err
		}

		fields = append(fields, field)
	}

	return fields, nil
}

func readSerialVersionUID(stream *ObjectStream) (int64, error) {
	bs, err := stream.ReadN(8)
	if err != nil {
		return 0, fmt.Errorf("read SerialVersionUID failed on index %v", stream.CurrentIndex())
	}

	// uint64 to int64 is expected
	return int64(binary.BigEndian.Uint64(bs)), nil
}