FROM golang AS build

WORKDIR /build
COPY ./go.mod ./minesweeper.go ./
RUN go build

FROM scratch AS final
COPY --from=build /build/minesweeper ./

CMD ["./minesweeper"]
