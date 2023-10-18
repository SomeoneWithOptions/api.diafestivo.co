# Colombian Holidays API

This is an API built using Go that provides information about holidays in Colombia.

## Table of Contents

- [Usage](#usage)
- [API Endpoints](#api-endpoints)

## Usage

To use this API, send a GET request to the desired endpoint. the API responds with a JSON object

## API Endpoints

### `/all`

Returns an array with all holidays for the year that are not sunday.
each item in the array has 2 properties : 
name: name of the hiliday in spanish
date : date of the holiday in ISO format 

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
]
```

### `/next`

Returns an object with information about the next holiday in colombia that is not Sunday 

**Response:**

- `200 OK` on success.

```json
{
	"name":"el Día de la Independencia",
	"date":"2023-07-20T00:00:00.000Z",
	"isToday":false,
	"daysUntil":15
}
```