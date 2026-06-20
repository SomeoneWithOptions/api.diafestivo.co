# Colombian Holidays API

API for Colombian public holidays.

All responses include `Access-Control-Allow-Origin: *`.

## Endpoints

### `GET /all`

Returns holidays for the current year in Colombia time.

```sh
curl -L https://api.diafestivo.co/all
```

```json
[
  {
    "date": "2025-01-01T00:00:00Z",
    "name": "Año Nuevo"
  }
]
```

### `GET /next`

Returns the next holiday.

```json
{
  "name": "Corpus Christi",
  "date": "2025-06-23T00:00:00Z",
  "isToday": false,
  "daysUntil": 13
}
```

### `GET /is/{date}`

Checks a date in `YYYY-MM-DD` format.

```sh
curl -L https://api.diafestivo.co/is/2025-01-01
```

```json
{
  "isHoliday": true
}
```

Invalid dates return `400 Bad Request` with body `error parsing date`.

### `GET /make?year=YYYY`

Returns holidays for the requested year.

```sh
curl -L 'https://api.diafestivo.co/make?year=2027'
```

```json
[
  {
    "date": "2027-01-01T00:00:00Z",
    "name": "Año Nuevo"
  },
  {
    "date": "2027-01-11T00:00:00Z",
    "name": "el Día de los Reyes Magos"
  }
]
```

Invalid years return `400 Bad Request` with body `error parsing year`.

### `GET /template`

Returns an HTML fragment for the current/next holiday.

### `GET /left`

Returns an HTML fragment with remaining holidays.

### `GET /healthz`

Returns `ok` for health checks.

### Invalid routes

Invalid routes return `400 Bad Request`:

```json
{
  "status": 400,
  "message": "Please Use Valid Routes:",
  "valid_routes": ["/all", "/next", "/is/YYYY-MM-DD", "/make?year=YYYY"]
}
```
