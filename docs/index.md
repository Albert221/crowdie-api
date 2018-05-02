# Groupie API documentation

Groupie API consists of endpoints serving for:

- creating new group
- fetching data from group
- joining existing group
- updating member's role
- pushing new *coords bits* to group
- kicking member from group

## Concepts

*Coords bit* - a pair of geocoordinates (latitude and longitude) and the time that those coordinates were sent.

*Security token* - a [JSON Web Token](https://jwt.io) containing user's android id and secret. This can be obtained from response when creating new group or joining the existing one.

*Lost member* - a member that was already present in the given group, but joined it again as if one was *lost*.

## Endpoints

All endpoints below reside in paths prefixed with `/api/v1` as it's the first version of API.

Endpoints can return a `500 Internal Server Error` apart from specified ones. In that case, the error has been logged with ERROR severity.

Besides that, endpoints accessing existing data (i.e. every endpoint except creating new group and joining existing group in specific case) require a security token to be present. If that's not met, a `401 Unauthorized` will be returned. This happens also when the given token is invalid.

### Creating new group

This endpoint creates new group with the only member created from supplied data. That member is an admin of the group. Endpoint returns group data in `group` element, as well as member's id in `yourId` and a `token` to be sent in each next request to that group.

#### Request

```http
POST /group
Content-Type: application/json

{
    "name": "John Doe",
    "androidId": "8a416d2cb454759a",
    "lat": 62.96398,
    "lng": 87.57387
}
```

#### Response

```http
HTTP/1.1 201 Created

{
    "group": {
        "id": "c0c83e12-5ebb-4626-bb8f-80778ef15b49",
        "members": [
            {
                "id": "acde282a-3926-4228-818d-8aa8657abfa8",
                "name": "John Doe",
                "role": 1,
                "coordsBit": {
                    "lat": 62.96398,
                    "lng": 87.57387,
                    "time": 1525283259
                }
            }
        ]
    },
    "yourId": "acde282a-3926-4228-818d-8aa8657abfa8",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6I..."
}
```

### Fetching data from group

This endpoints returns all public data of specified group. It returns `404 Not Found` if given group does not exist and `403 Forbidden` if user with given security token does not belong to the given group.

#### Request

```http
GET /group/c0c83e12-5ebb-4626-bb8f-80778ef15b49
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6I...
```

#### Response

```http
HTTP/1.1 200 OK

{
    "id": "c0c83e12-5ebb-4626-bb8f-80778ef15b49",
    "members": [
        {
            "id": "acde282a-3926-4228-818d-8aa8657abfa8",
            "name": "John Doe",
            "role": 1,
            "coordsBit": {
                "lat": 62.96398,
                "lng": 87.57387,
                "time": 1525283259
            }
        },
        // (...)
    ]
}
```

### Joining existing group

Joining existing group can happen in one of two cases:

- user has never been in the given group and joins it
- user has already been in the given group, but something went wrong and he is joining it again (he was a *lost member*)

In the first case, situation is almost exactly same as with creating new group, but the difference is that we specify the group id. Same response is returned as in the mentioned endpoint.

In the second case, if the user has already been a member of given group (i.e. one's android id was present in group's members' android ids), a valid security token of this user must be provided. Otherwise, an `401 Unauthorized` or `403 Forbidden` will be returned, respectively to the lack of token or the token of wrong user.

This endpoint will return a `404 Not Found` when the given group does not exist.

#### Request for first case

```http
POST /group/c0c83e12-5ebb-4626-bb8f-80778ef15b49/member
Content-Type: application/json

{
    "name": "John Doe",
    "androidId": "8a416d2cb454759a",
    "lat": 62.96398,
    "lng": 87.57387
}
```

#### Request for second case

```http
POST /group/c0c83e12-5ebb-4626-bb8f-80778ef15b49/member
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6I...

{
    "name": "John Doe",
    "androidId": "8a416d2cb454759a",
    "lat": 62.96398,
    "lng": 87.57387
}
```

#### Response

In the second case, `yourId` and `token` keys will not be returned.

```http
HTTP/1.1 201 Created

{
    "group": {
        "id": "c0c83e12-5ebb-4626-bb8f-80778ef15b49",
        "members": [
            {
                "id": "acde282a-3926-4228-818d-8aa8657abfa8",
                "name": "John Doe",
                "role": 1,
                "coordsBit": {
                    "lat": 62.96398,
                    "lng": 87.57387,
                    "time": 1525283259
                }
            }
        ]
    },
    "yourId": "acde282a-3926-4228-818d-8aa8657abfa8",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6I..."
}
```

### Updating member's role

This endpoint is for changing given member's role.

If the user providing given security token is not an admin, a `403 Forbidden` will be returned. If given member does not exist, a `404 Not Found` will be returned.

#### Request

```http
PATCH /member/acde282a-3926-4228-818d-8aa8657abfa8/role
Content-Type: application/x-www-form-urlencoded
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6I...

role=1
```

#### Response

Just the group with updated member's role, same as the `group` item of response above.

### Pushing new *coords bits* to group

This endpoint updates given member's coords bit with given latitude and longitude and sets the coords bit's time to now.

If the user is trying to push coords bit not to himself, a `403 Forbidden` will be returned. If given member does not exist, a `404 Not Found` will be returned.

#### Request

```http
PATCH /member/acde282a-3926-4228-818d-8aa8657abfa8/coords-bit
Content-Type: application/x-www-form-urlencoded
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6I...

lat=62.96398&lng=87.57387
```

#### Response

Just the group with updated member's coords bit, same as the `group` item of two responses above.

### Kicking member from group

Endpoint for kicking members from group or leaving it. It can be ran by an admin to kick anyone from the given group or by ordinary user to kick himself (leave).

If the given member (to be kicked) is the one of given security token, one will be kicked. If the given member is an other member:

- a `403 Forbidden` will be returned if given security token's user is not an admin
- the given member will be kicked if given security token's user is an admin.

If the given member does not exist, a `404 Not Found` will be returned.

#### Request

```http
DELETE /member/acde282a-3926-4228-818d-8aa8657abfa8
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6I...
```

#### Response

Just the group without kicked member, same as the `group` item of three responses above.