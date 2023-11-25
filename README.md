# Nano Flow

Tiny append only database used for event driven applications. Uses Grpc to produce messages and subscribe to messages via ephemeral consumer groups.



## Message Storage Protocol

![nano-flow-file-protocol](https://github.com/redsuperbat/nano-flow/assets/44667637/a9d9b4f0-e357-4442-992a-8218aa7b6e3e)

2 bytes

Version Nr as a unsigned 16 bit integer

4 bytes 

Message content length as a unsigned 32 bit integer

8 bytes

Timestamp of message as a signed 64 bit integer

4 bytes

CRC32 checksum as a 32 bit unsigned integer. Checksum includes the header and body

19-X bytes

Message body defined by the content length in header



