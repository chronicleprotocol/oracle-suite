PB = $(wildcard *.proto)
GO = $(PB:.proto=.pb.go)

all: $(GO)

%.pb.go: %.proto
		protoc --go_out=. --go_opt=paths=source_relative $<

clean:
		rm -f *.pb.go
		rm -f *.go
