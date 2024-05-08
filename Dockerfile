FROM golang AS build

WORKDIR /build
COPY ./ ./
RUN go build

FROM scratch AS final
COPY --from=build /build/go-minesweeper ./

CMD ["./go-minesweeper"]
