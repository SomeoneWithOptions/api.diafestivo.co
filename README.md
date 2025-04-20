# Colombian Holidays API

This is an API built using Go that provides information about holidays in Colombia.

## Table of Contents

- [Usage](#usage)
- [API Endpoints](#api-endpoints)

## Usage

To use this API, send a GET request to the desired endpoint. the API responds with a JSON object

## API Endpoints

### `/all`

Returns an array with all holidays for the year.
Each item in the array has 2 properties:
name: name of the hiliday in spanish
date : date of the holiday in ISO format

**Example:**

`curl -L api.diafestivo.co/all`

**Response:**

- `200 OK` on success.

```json
[
	{
		"name": "Día de los Reyes Magos",
		"date": "2023-01-09T00:00:00.000Z"
	},
	{
		"name": "Día de San José",
		"date": "2023-03-20T00:00:00.000Z"
	},
	{
		"name": "Jueves Santo",
		"date": "2023-04-06T00:00:00.000Z"
	},
    ...
]
```

### `/next`

Returns an object with information about the next holiday in colombia that is NOT Sunday

**Example:**

`curl -L api.diafestivo.co/next`

**Response:**

- `200 OK` on success.

```json
{
    "name": "el Día de la Independencia",
    "date": "2023-07-20T00:00:00.000Z",
    "isToday": false,
    "daysUntil": 15
}
```

### `/is/(date)`

Make a GET request to "/is/{date}", where "{date}" represents a date you want to check if it is a colombian holiday in the format "YYYY-MM-DD".

**Example:**

`curl -L api.diafestivo.co/is/2025-06-11`

**Response:**

- `200 OK` on success.

```json
{
    "is_holiday": false
}
```

### `/make?year=YYYY`

Returns an array with all holidays for the year requested

**Example:**

`curl -L api.diafestivo.co/make?year=2027`

**Response:**

- `200 OK` on success.

```json
[
  {
    "name": "2027-01-01T00:00:00Z",
    "date": "Año Nuevo"
  },
  {
    "name": "2027-01-11T00:00:00Z",
    "date": "el Día de los Reyes Magos"
  },
  {
    "name": "2027-03-22T00:00:00Z",
    "date": "el Día de San José"
  },
...
]
```
