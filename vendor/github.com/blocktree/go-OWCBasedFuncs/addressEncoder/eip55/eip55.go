package eip55

import(
	"errors"
	"encoding/hex"
	"strings"
	"github.com/blocktree/go-OWCrypt"
)
var (
	ErrorInvalidAddress  = errors.New("Invalid address!")
)

func byte_to_half(in[]byte, in_len int)[]byte{
	out := make([]byte, in_len<<1)
for i:=0;i<in_len;i++{
	out[2*i+0]=(in[i]>>4)&0x0f
	out[2*i+1]=in[i]&0x0f
}
return out
}

func hex_to_str(inchar[]uint8,inchar_len int)[]byte{
	hbit:=byte(1)
	lbit:=byte(1)
	out :=make([]byte,inchar_len<<1)
	
	i:=0
	for ;i<inchar_len;i++{
	hbit=(inchar[i]&0xf0)>>4
	lbit=inchar[i]&0x0f
    if hbit>9{
     out[2*i]='a'+hbit-10
	}else{
		out[2*i]='0'+hbit
	}

	if lbit>9{
		out[2*i+1]='a'+lbit-10
	}else{
		out[2*i+1]='0'+lbit
	}
	}
	//out[2*i]=0
	return out
}


func Eip55_encode(addr[]byte)string{
	encode_addr :=make([]byte,40)
	addr_hex:=make([]byte,40)
	knecck256:=make([]byte,32)
	knecck256_hex:=make([]byte,64)
	addr_hex=hex_to_str(addr[:],20)
	knecck256=owcrypt.Hash(addr_hex, 0, owcrypt.HASH_ALG_KECCAK256)
	knecck256_hex=byte_to_half(knecck256[:], 32)
	for i:=0;i<40;i++{
		if((addr_hex[i]>=48)&&(addr_hex[i]<=57)){
			encode_addr[i]=addr_hex[i]
		}else{
			if knecck256_hex[i]>=8{
				encode_addr[i]=addr_hex[i]-32
			}else{
				encode_addr[i]=addr_hex[i]
			}
		}
	}
	str:=string(encode_addr)
	return str
}

func Eip55_decode(encode_addr string)([]byte,error){
	decode_addr:=make([]byte,20)
	var check_addr string
	for _,ch:=range encode_addr{
		if (ch < 48)||((ch>57) &&(ch<65))||((ch>70)&&(ch<97))||(ch>102){
			return nil,ErrorInvalidAddress
		} 
	}
	decode_addr,err:= hex.DecodeString(encode_addr[:])
	if err!=nil{

	   return nil,err
	}
	check_addr=Eip55_encode(decode_addr[:])
	ret:=strings.Compare(encode_addr,check_addr)
	if ret==-1{
		return nil,ErrorInvalidAddress
	}
    return decode_addr,nil
}

