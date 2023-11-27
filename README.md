# Trancestry
Gives ancestor count for a given block

Build : go build trancestry.go 

Run : ./trancestry <block_number>

sample run

gmanojkumar@C2VZP1N3MD6N trancestry % ./trancestry 
panic: block info not provided

goroutine 1 [running]:
main.main()
	/Users/gmanojkumar/Workspace/src/trancestry/trancestry.go:35 +0x1d8
gmanojkumar@C2VZP1N3MD6N trancestry % ./trancestry 60
fetching transaction map for block:  60
hashrespmap:  00000000c47a58b89434bd7c7af8334b366f713330dfb89119d67d79424d2cdf
total transactions:  [{82a1e1731a9b22fdea55c09f2fac191a89efee127956fcfef65caab70f54002d [0000000000000000000000000000000000000000000000000000000000000000]}]
transaction map:  map[82a1e1731a9b22fdea55c09f2fac191a89efee127956fcfef65caab70f54002d:0]
top10 transactions for block:  [{82a1e1731a9b22fdea55c09f2fac191a89efee127956fcfef65caab70f54002d 0}]
gmanojkumar@C2VZP1N3MD6N trancestry % 
