# Colombian Holidays API

API for Colombian public holidays.

Base URL: `https://api.diafestivo.co`

All responses include `Access-Control-Allow-Origin: *`.

## Endpoints

### `GET /all`

Returns all holidays for the current year in Colombia time.

```sh
curl -L https://api.diafestivo.co/all
```

Example response (first 4 holidays):

```json
[
  {
    "date": "2026-01-01T00:00:00Z",
    "name": "Año Nuevo"
  },
  {
    "date": "2026-01-12T00:00:00Z",
    "name": "el Día de los Reyes Magos"
  },
  {
    "date": "2026-03-23T00:00:00Z",
    "name": "el Día de San José"
  },
  {
    "date": "2026-04-02T00:00:00Z",
    "name": "Jueves Santo"
  }
]
```

### `GET /next`

Returns the next holiday from today's date in Colombia time.

```sh
curl -L https://api.diafestivo.co/next
```

```json
{
  "name": "Corpus Christi",
  "date": "2025-06-23T00:00:00Z",
  "isToday": false,
  "daysUntil": 13
}
```

### `GET /is/{date}`

Checks if a date is a holiday. Use `YYYY-MM-DD` format.

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

Returns all holidays for the requested year.

```sh
curl -L 'https://api.diafestivo.co/make?year=2027'
```

Example response (first 4 holidays):

```json
[
  {
    "date": "2027-01-01T00:00:00Z",
    "name": "Año Nuevo"
  },
  {
    "date": "2027-01-11T00:00:00Z",
    "name": "el Día de los Reyes Magos"
  },
  {
    "date": "2027-03-22T00:00:00Z",
    "name": "el Día de San José"
  },
  {
    "date": "2027-03-25T00:00:00Z",
    "name": "Jueves Santo"
  }
]
```

Invalid years return `400 Bad Request` with body `error parsing year`.

### Invalid routes

Invalid routes return `400 Bad Request`:

```json
{
  "status": 400,
  "message": "Please Use Valid Routes:",
  "valid_routes": ["/all", "/next", "/is/YYYY-MM-DD", "/make?year=YYYY"]
}
```
