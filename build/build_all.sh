GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc CXX=i686-w64-mingw32-g++ go build -ldflags="-w -s" -o spyac_win_i386.exe ../cmd/spyac/spyac.go
GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc CXX=i686-w64-mingw32-g++ go build -ldflags="-w -s" -o spyad_win_i386.exe ../cmd/spyad/spyad.go
GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc CXX=i686-w64-mingw32-g++ go build -ldflags="-w -s" -o spyar_win_i386.exe ../cmd/spyar/spyar.go

GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ go build -ldflags="-w -s" -o spyac_win_amd64.exe ../cmd/spyac/spyac.go
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ go build -ldflags="-w -s" -o spyad_win_amd64.exe ../cmd/spyad/spyad.go
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ go build -ldflags="-w -s" -o spyar_win_amd64.exe ../cmd/spyar/spyar.go

CC=gcc CXX=g++ CGO_ENABLED=1 GOARCH=386 go build -ldflags="-w -s" -o spyac_nix_i386 ../cmd/spyac/spyac.go
CC=gcc CXX=g++ CGO_ENABLED=1 GOARCH=386 go build -ldflags="-w -s" -o spyad_nix_i386 ../cmd/spyad/spyad.go
CC=gcc CXX=g++ CGO_ENABLED=1 GOARCH=386 go build -ldflags="-w -s" -o spyar_nix_i386 ../cmd/spyar/spyar.go

CC=gcc CXX=g++ CGO_ENABLED=1 GOARCH=amd64 go build -ldflags="-w -s" -o spyac_nix_amd64 ../cmd/spyac/spyac.go
CC=gcc CXX=g++ CGO_ENABLED=1 GOARCH=amd64 go build -ldflags="-w -s" -o spyad_nix_amd64 ../cmd/spyad/spyad.go
CC=gcc CXX=g++ CGO_ENABLED=1 GOARCH=amd64 go build -ldflags="-w -s" -o spyar_nix_amd64 ../cmd/spyar/spyar.go
