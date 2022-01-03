# Convert Youtube links to GIF files

## Usage

```bash
go build ; ./go-youtube-gif-converter.exe

```

## API examples

### `POST http://localhost:8080/gif`

#### Request Body
```
{
    "videoID": "3PYC_SjzKsA"
}
```

#### Response Body

```json
{
    "conversionUUID": "6fb3bcde-3067-44ad-ae2d-51262b30ed81"
}
```

### `GET http://localhost:8080/gif/3PYC_SjzKsA.gif`

![# GIF as binary data](/example.gif)

### `GET http://localhost:8080/gif/status/3PYC_SjzKsA`

```json
[
    {
        "conversionUUID": "03401191-f697-4607-9eba-7d551f6c569a",
        "status": "pending",
        "videoID": "3PYC_SjzKsA"
    },
    {
        "conversionUUID": "39d44faa-0f27-4737-91ed-5105c2dba268",
        "status": "pending",
        "videoID": "3PYC_SjzKsA"
    },
    {
        "conversionUUID": "4e5e4cfe-76c3-4733-9674-4eb4f024ff6f",
        "status": "done",
        "videoID": "3PYC_SjzKsA"
    },
    {
        "conversionUUID": "a06cc4fc-793c-4dfa-962e-90bfe36d6b01",
        "status": "converting",
        "videoID": "3PYC_SjzKsA"
    },
    {
        "conversionUUID": "3cc04e63-2cb5-45a1-a52e-1f5587fc7a85",
        "status": "done",
        "videoID": "3PYC_SjzKsA"
    }
]
```