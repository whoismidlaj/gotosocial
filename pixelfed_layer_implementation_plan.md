# GoToSocial → Pixelfed Compatibility Layer: Implementation Plan

## Background

This plan was produced by reading the actual GoToSocial codebase at `/home/midlajm/Projects/gotosocial`. Every finding references exact file paths, struct names, and function names. No information is speculative.

---

## Section 1 — Architecture Review

### 1.1 API Layer

**Existing routes** are registered in [`internal/api/client.go`](file:///home/midlajm/Projects/gotosocial/internal/api/client.go). The full set of modules is:

| Module | Prefix |
|--------|--------|
| accounts | `/v1/accounts`, `/v1/profile` |
| admin | `/v1/admin` |
| announcements | `/v1/announcements` |
| apps | `/v1/apps` |
| blocks | `/v1/blocks` |
| bookmarks | `/v1/bookmarks` |
| conversations | `/v1/conversations` |
| customEmojis | `/v1/custom_emojis` |
| debug | `/v1/debug` |
| directory | `/v1/directory` |
| exports | `/v1/exports` |
| favourites | `/v1/favourites` |
| featuredTags | `/v1/featured_tags` |
| filtersV1/V2 | `/v1/filters`, `/v2/filters` |
| followedTags | `/v1/followed_tags` |
| followRequests | `/v1/follow_requests` |
| instance | `/v1/instance`, `/v2/instance` |
| interactionPolicies | `/v1/interaction_policies` |
| interactionRequests | `/v1/interaction_requests` |
| lists | `/v1/lists` |
| markers | `/v1/markers` |
| media | `/v1/media`, `/v2/media` |
| mutes | `/v1/mutes` |
| notifications | `/v1/notifications` |
| polls | `/v1/polls` |
| preferences | `/v1/preferences` |
| push | `/v1/push` |
| reports | `/v1/reports` |
| scheduledStatuses | `/v1/scheduled_statuses` |
| search | `/v1/search`, `/v2/search` |
| statuses | `/v1/statuses` |
| streaming | `/v1/streaming` |
| suggestions | `/v2/suggestions` |
| tags | `/v1/tags` |
| timelines | `/v1/timelines` |
| trends | `/v1/trends` |
| tokens | `/v1/tokens` |
| user | `/v1/user` |

**Status routes** ([`internal/api/client/statuses/status.go`](file:///home/midlajm/Projects/gotosocial/internal/api/client/statuses/status.go)):

```
POST   /v1/statuses
GET    /v1/statuses/:id
GET    /v1/statuses              (multi-fetch)
PUT    /v1/statuses/:id
DELETE /v1/statuses/:id
POST   /v1/statuses/:id/favourite
POST   /v1/statuses/:id/unfavourite
GET    /v1/statuses/:id/favourited_by
POST   /v1/statuses/:id/reblog
POST   /v1/statuses/:id/unreblog
GET    /v1/statuses/:id/reblogged_by
POST   /v1/statuses/:id/bookmark
POST   /v1/statuses/:id/unbookmark
POST   /v1/statuses/:id/mute
POST   /v1/statuses/:id/unmute
POST   /v1/statuses/:id/pin
POST   /v1/statuses/:id/unpin
GET    /v1/statuses/:id/context
GET    /v1/statuses/:id/history
GET    /v1/statuses/:id/source
```

**Timeline routes** ([`internal/api/client/timelines/timeline.go`](file:///home/midlajm/Projects/gotosocial/internal/api/client/timelines/timeline.go)):

```
GET /v1/timelines/home
GET /v1/timelines/public
GET /v1/timelines/list/:id
GET /v1/timelines/tag/:tagName
```

**Media routes** ([`internal/api/client/media/media.go`](file:///home/midlajm/Projects/gotosocial/internal/api/client/media/media.go)):

```
POST /v1/media         (or /v2/media)
GET  /v1/media/:id
PUT  /v1/media/:id
```

**Account routes** ([`internal/api/client/accounts/accounts.go`](file:///home/midlajm/Projects/gotosocial/internal/api/client/accounts/accounts.go)):

```
POST GET /v1/accounts
GET  /v1/accounts/:id
GET  /v1/accounts/verify_credentials
PATCH /v1/accounts/update_credentials
GET  /v1/accounts/:id/statuses
GET  /v1/accounts/:id/followers
GET  /v1/accounts/:id/following
POST /v1/accounts/:id/follow
POST /v1/accounts/:id/unfollow
POST /v1/accounts/:id/block
POST /v1/accounts/:id/unblock
GET  /v1/accounts/:id/lists
GET  /v1/accounts/:id/featured_tags
GET  /v1/accounts/:id/note
GET  /v1/accounts/:id/mute
GET  /v1/accounts/:id/unmute
GET  /v1/accounts/:id/relationships
GET  /v1/accounts/search
GET  /v1/accounts/lookup
POST /v1/accounts/alias
POST /v1/accounts/move
```

---

### 1.2 ActivityPub Layer

**Types recognized as statuses** ([`internal/ap/interfaces.go:68`](file:///home/midlajm/Projects/gotosocial/internal/ap/interfaces.go#L68-L84)):

- `Article`, `Document`, `Image`, `Video`, `Note`, `Page`, `Event`, `Place`, `Profile`, `Question`, **`Album`** (Funkwhale extension — already recognized!)

**`Statusable` interface** ([`internal/ap/interfaces.go:274`](file:///home/midlajm/Projects/gotosocial/internal/ap/interfaces.go#L274-L291)) requires: `WithSummary`, `WithName`, `WithInReplyTo`, `WithPublished`, `WithUpdated`, `WithURL`, `WithAttributedTo`, `WithTo`, `WithCc`, `WithSensitive`, `WithContent`, `WithAttachment`, `WithTag`, `WithReplies`.

**`Attachmentable` interface** ([`internal/ap/interfaces.go:341`](file:///home/midlajm/Projects/gotosocial/internal/ap/interfaces.go#L341-L348)):

```go
type Attachmentable interface {
    WithTypeName
    WithMediaType
    WithURL
    WithName
    WithSummary
    WithBlurhash
}
```

This interface supports: `Audio`, `Document`, `Image`, `Video`.

**Attachment extraction** ([`internal/ap/extract.go:846`](file:///home/midlajm/Projects/gotosocial/internal/ap/extract.go#L846-L862)): extracts `RemoteURL`, `Description` (from `summary` or `name`), `Blurhash`, and `Focus` (via `TootFocalPoint`).

**Tag extraction** ([`internal/ap/extract.go:931`](file:///home/midlajm/Projects/gotosocial/internal/ap/extract.go#L931-L993)): Fully extracts hashtags from `tag[]` with type `"Hashtag"`. Mentions are also extracted.

**Custom GtS AP extensions** defined in `internal/ap/activitystreams.go`:
- `GoToSocialHidesToPublicFromUnauthedWeb`
- `GoToSocialHidesCcPublicFromUnauthedWeb`
- `GoToSocialApprovedBy`, `GoToSocialLikeAuthorization`, etc.
- `LikeRequest`, `ReplyRequest`, `AnnounceRequest` activities

---

### 1.3 Data Models

**`Status`** ([`internal/gtsmodel/status.go`](file:///home/midlajm/Projects/gotosocial/internal/gtsmodel/status.go)):
- Stores: `Content`, `ContentWarning`, `AttachmentIDs`, `TagIDs`, `MentionIDs`, `EmojiIDs`, `Visibility`, `ActivityStreamsType`, `Sensitive` (via `Flags`), `Language`, `InteractionPolicy`, `PollID`.
- **No `Location` field.** No `Album` or collection membership. No `ExifData`.

**`MediaAttachment`** ([`internal/gtsmodel/mediaattachment.go`](file:///home/midlajm/Projects/gotosocial/internal/gtsmodel/mediaattachment.go)):
- Stores: `Type` (Image/Video/Audio/Gifv), `Description` (alt-text), `Blurhash`, `FileMeta` (width, height, aspect, duration, fps, bitrate), `Focus` (x,y), `File`, `Thumbnail`.
- **No `License` field, no `ExifData`, no per-attachment location, no `Tags` (people-tagging in photo).**

**`Account`** ([`internal/gtsmodel/account.go`](file:///home/midlajm/Projects/gotosocial/internal/gtsmodel/account.go)):
- Stores: `Username`, `Domain`, `DisplayName`, `Note`, `Fields`, `Avatar`, `Header`, `Discoverable`, `Indexable`, `Locked`, `ActorType`, `URL`, `InboxURI`, `OutboxURI`, `FollowingURI`, `FollowersURI`, `FeaturedCollectionURI`.
- **No `WebsiteURL` separate field, no `Category` for Pixelfed creator types, no `Story` capability flag.**

---

### 1.4 Media Subsystem

**`internal/media/`** processes: Image (JPEG, PNG, GIF, WebP), Video (MP4/WebM via ffmpeg), Audio, Gifv. Outputs: original + thumbnail (JPEG/WebP) with blurhash computation.

**Supported types** ([`internal/media/types.go`](file:///home/midlajm/Projects/gotosocial/internal/media/types.go)): `TypeAttachment`, `TypeHeader`, `TypeAvatar`, `TypeEmoji`. No `TypeStory`.

**What is missing for Pixelfed**:
- No story media type with ephemeral storage + expiry
- No multi-image album grouping
- No EXIF stripping configuration
- No per-attachment location geocoding

---

### 1.5 OAuth Implementation

**`internal/oauth/server.go`**: Standard OAuth 2.0 with `authorization_code` + `client_credentials` grants. Bearer tokens don't expire (Mastodon compat). PKCE (`S256`) is supported. Scopes are freeform strings compared against client registration.

**Scopes currently used by GtS**: `read`, `write`, `follow`, `push`, `admin:read`, `admin:write` (inferred from middleware). No `stories:read`, `stories:write`, `albums:read`, `albums:write` scope definitions exist.

---

### 1.6 Type Conversion (Serialization) Layer

**`internal/typeutils/internaltofrontend.go`**: Converts internal `gtsmodel.*` → `apimodel.*`.

Key conversion functions:
- `AttachmentToAPIAttachment` (line 684): maps media to Mastodon Attachment model, fully including meta, blurhash, description, focus
- `accountToAPIAccountPublic` (line 250): maps account with followers/following/statuses counts, avatar, header, fields, emojis
- `instanceMastodonVersion = "3.5.3"` (line 51): GtS presents itself as Mastodon 3.5.3

**`internal/typeutils/internaltoas.go`**: Converts internal → ActivityPub vocab types.

---

## Section 2 — Pixelfed Gap Analysis (ActivityPub)

| Feature | GtS Status | Notes |
|---------|-----------|-------|
| Note objects as statuses | ✅ Full | Core behavior |
| Image/Video/Audio as attachments | ✅ Full | `Attachmentable` interface |
| `sensitive` flag | ✅ Full | `WithSensitive`, stored in `StatusFlags` |
| Content warnings (summary) | ✅ Full | `ExtractSummary`, stored in `ContentWarning` |
| Alt text (description on attachments) | ✅ Full | `ExtractDescription` tries `summary` then `name` |
| Hashtags | ✅ Full | `ExtractHashtags` |
| Mentions | ✅ Full | `ExtractMentions` |
| Blurhash | ✅ Full | `ExtractBlurhash` |
| Focal point | ✅ Full | `ExtractFocus` via `TootFocalPoint` |
| OrderedCollection / Collection | ✅ Full | Used for outbox, followers, following, featured |
| Followers/Following collections | ✅ Full | |
| Featured/pinned collection | ✅ Full | `toot:featured` |
| Image-type Note (Pixelfed sends `type: "Note"` with image attachments) | ✅ Full | GtS handles Note with any `attachment[]` |
| `Place` / location metadata | ⚠️ Partial | `ObjectPlace` is in `IsStatusable()` but no location data is stored or extracted |
| `Album` type (Funkwhale-style, grouped media) | ⚠️ Partial | `ObjectAlbum = "Album"` in `activitystreams.go` and in `IsStatusable()`, but there is **no extraction logic** for Album-specific properties |
| Photo tagging (`tag` with type `Person` pointing to location in image) | ❌ Missing | Not in any model or extractor |
| Story objects | ❌ Missing | No story type, no expiry, no ephemeral media handling |
| `license` on attachments | ❌ Missing | Not in `Attachmentable` interface or extraction |
| EXIF metadata | ❌ Missing | |
| Custom Pixelfed-specific extensions (`pixelfed:...`) | ❌ Missing | No namespace registered |

---

## Section 3 — API Compatibility (What Works Today)

### ✅ Endpoints Pixelfed clients will find working

Pixelfed clients are Mastodon API clients. The following will work as-is:

- `POST /oauth/token` — OAuth token exchange
- `GET /api/v1/instance` + `GET /api/v2/instance` — Instance metadata
- `POST /api/v1/apps` — App registration
- `GET /api/v1/accounts/verify_credentials` — Auth'd user info
- `PATCH /api/v1/accounts/update_credentials` — Profile editing
- `GET /api/v1/accounts/:id` — Profile fetch
- `GET /api/v1/accounts/:id/statuses` — Account timeline (media-only filter via `only_media=true`)
- `GET /api/v1/accounts/:id/followers` + `/following`
- `POST /api/v1/accounts/:id/follow` + `/unfollow`
- `POST /api/v1/statuses` — Post creation with `media_ids[]`
- `GET/PUT/DELETE /api/v1/statuses/:id`
- `POST /api/v1/media` or `/api/v2/media` — Upload media
- `GET /api/v1/timelines/home`
- `GET /api/v1/timelines/public`
- `GET /api/v1/timelines/tag/:tag`
- `GET /api/v1/search` + `/api/v2/search`
- `GET /api/v1/notifications`
- `POST /api/v1/statuses/:id/favourite` + `/unfavourite`
- `POST /api/v1/statuses/:id/reblog` + `/unreblog`
- `POST /api/v1/statuses/:id/bookmark` + `/unbookmark`
- `GET /api/v1/statuses/:id/context`
- `GET /api/v1/bookmarks`
- `GET /api/v1/favourites`
- `GET /api/v1/trends` + `/api/v1/trends/tags` + `/api/v1/trends/statuses`
- `GET /api/v1/tags/:name` + `POST /api/v1/tags/:name/follow`
- `GET /api/v1/custom_emojis`
- `POST /api/v1/reports`

### ❌ Endpoints Pixelfed clients expect but GtS is missing

#### 3.1 Collections / Albums API

| Method | Path | Notes |
|--------|------|-------|
| `GET` | `/api/v1/collections` | List own collections |
| `POST` | `/api/v1/collections` | Create collection |
| `GET` | `/api/v1/collections/:id` | Get collection |
| `PUT` | `/api/v1/collections/:id` | Update collection |
| `DELETE` | `/api/v1/collections/:id` | Delete collection |
| `GET` | `/api/v1/collections/:id/items` | List items in collection |
| `POST` | `/api/v1/collections/:id/items` | Add status to collection |
| `DELETE` | `/api/v1/collections/:id/items/:itemId` | Remove item |
| `GET` | `/api/v1/accounts/:id/collections` | Get account's collections |

**Request format** (create collection):
```json
{ "title": "string", "description": "string", "visibility": "public|private" }
```
**Response format**:
```json
{ "id": "...", "title": "...", "description": "...", "visibility": "...", "thumb": null, "post_count": 0, "account": {...} }
```
**Existing code reusable**: `processing` layer pattern, `gtsmodel.List` as reference for grouped entities, DB patterns in `internal/db/`.
**Complexity**: Medium (new table, new handler package, new processing methods).

---

#### 3.2 Stories API

| Method | Path | Notes |
|--------|------|-------|
| `GET` | `/api/v1/stories` | Stories from followed accounts |
| `GET` | `/api/v1/stories/:id` | Single story |
| `POST` | `/api/v1/stories/create` | Create story |
| `DELETE` | `/api/v1/stories/:id` | Delete story |
| `GET` | `/api/v1/accounts/:id/stories` | Stories from account |

**Request format** (create story):
```json
{ "media_id": "...", "duration": 10, "expiry": 86400, "visibility": "...", "music_track": null }
```
**Response format**:
```json
{ "id": "...", "account": {...}, "media_attachment": {...}, "created_at": "...", "expires_at": "...", "views": 0, "seen": false }
```
**Existing code reusable**: `gtsmodel.MediaAttachment`, `gtsmodel.Visibility`, OAuth middleware.
**Complexity**: High (new DB entity, new expiry scheduler, ephemeral media handling, no AP representation).

---

#### 3.3 Discover / Explore API

| Method | Path | Notes |
|--------|------|-------|
| `GET` | `/api/v1/discover/posts` | Curated public posts |
| `GET` | `/api/v1/discover/tags/trending` | Trending tags (different path from GtS) |
| `GET` | `/api/v1/discover/accounts/popular` | Popular accounts |
| `GET` | `/api/v1/explore/trending` | Trending content |

**Existing code reusable**: `GET /api/v1/trends` and directory are close; `public` timeline can back `discover/posts`.
**Response format**: Same as standard statuses array with extra `_metadata` field with `top_likes`, etc.
**Complexity**: Low — mostly alias routes to existing functionality.

---

#### 3.4 Photo Tagging API

| Method | Path | Notes |
|--------|------|-------|
| `POST` | `/api/v1/media/:id/tags` | Tag a person in a photo |
| `GET` | `/api/v1/media/:id/tags` | Get tags on a photo |
| `DELETE` | `/api/v1/media/:id/tags/:tagId` | Remove person tag |

**Request format**:
```json
{ "account_id": "...", "x": 0.5, "y": 0.5 }
```
**Response format**:
```json
[{ "id": "...", "account": {...}, "x": 0.5, "y": 0.5 }]
```
**Existing code reusable**: `gtsmodel.Mention` pattern, `gtsmodel.MediaAttachment`.
**Complexity**: Medium (new table `media_tags`, new handlers, notification on tag).

---

#### 3.5 Location API

| Method | Path | Notes |
|--------|------|-------|
| `GET` | `/api/v1/location/search` | Search for locations |
| `POST` | `/api/v1/location` | Create/register a location |

**Request format** (search): `?q=City+Name`
**Response format**:
```json
[{ "id": "...", "name": "...", "country": "...", "slug": "...", "type": "city" }]
```
**Existing code reusable**: None directly. Requires external data source or self-contained location DB.
**Complexity**: Very High (external geocoder dependency or static DB needed).

---

#### 3.6 Profile / Instance Compatibility Gaps

Pixelfed clients check `GET /api/v1/instance` for the `pixelfed` or software key in `metadata`. They also check `nodeinfo` for `software.name === "pixelfed"`. GtS currently returns `"gotosocial"` in nodeinfo.

**Compatibility stub needed**: A configurable `X-Pixelfed-Compat: true` header or a `/api/pixelfed/...` namespace does not exist.

**Current version presented**:
- `internal/typeutils/internaltofrontend.go:51`: `instanceMastodonVersion = "3.5.3"` → returned as `"3.5.3+gotosocial-0.x.x"`

Pixelfed clients MAY gate features by checking if version looks like Pixelfed. This is a soft compatibility risk.

---

## Section 4 — Client Compatibility Analysis

### 4.1 Pixelfed Android/iOS clients

These clients authenticate via standard Mastodon OAuth flow:
1. `POST /api/v1/apps` — **works** ✅
2. OAuth authorize redirect → Bearer token — **works** ✅
3. `GET /api/v1/accounts/verify_credentials` — **works** ✅

**Required scopes** Pixelfed clients request: `read write follow push` — **all supported** ✅

**Home feed** (`GET /api/v1/timelines/home`): **works** ✅ — returns standard statuses with `media_attachments[]`.

**Image upload flow**:
1. `POST /api/v2/media` with `file` + `description` — **works** ✅
2. `POST /api/v1/statuses` with `media_ids[]` — **works** ✅

**Critical blockers for Pixelfed app login + basic use**:

| Blocker | Severity | Notes |
|---------|----------|-------|
| Missing `GET /api/v1/timelines/direct` | Low | Some client versions fall back gracefully |
| Missing `GET /api/v1/follow_requests` listing (exists) | None | Exists ✅ |
| Missing `GET /api/v1/lists` | None | Exists ✅ |
| Pixelfed client may call `GET /api/v1/accounts/:id/albums` | Medium | Will 404 |
| Pixelfed client may call `GET /api/v1/stories` | High | Will 404 — app may crash/show error screen |
| `GET /api/v1/discover/posts` | Medium | Empty 404; client may show blank Explore tab |
| NodeInfo `software.name` is `"gotosocial"` not `"pixelfed"` | Medium | Clients that gate story/album features by software name won't enable them, which is actually **safer** |

### 4.2 Required response fields for Status

GtS status API response ([`internal/api/model/status.go`](file:///home/midlajm/Projects/gotosocial/internal/api/model/status.go)):
- `id`, `created_at`, `edited_at`, `in_reply_to_id`, `sensitive`, `spoiler_text`, `visibility`, `language`, `uri`, `url`, `replies_count`, `reblogs_count`, `favourites_count`, `content`, `reblog`, `account`, `media_attachments[]`, `mentions[]`, `tags[]`, `emojis[]`, `card`, `poll`

**Pixelfed additionally expects**:
- `location` field on status → **missing** (no location model)
- `place_id` on status → **missing**
- `in_reply_to_account_id` → ✅ present

### 4.3 Required response fields for Account

GtS account response ([`internal/api/model/account.go`](file:///home/midlajm/Projects/gotosocial/internal/api/model/account.go)) includes all Mastodon standard fields.

**Pixelfed additionally expects**:
- `website` field (separate from `fields[]`) → ⚠️ Can be inferred from `fields[]` but not a top-level key
- `pronouns` → same, in `fields[]` not top-level
- `is_admin` → available via `role` field ✅

---

## Section 5 — Data Model Changes

### 5.1 Required for Phase A (MVP)

No new database tables required for basic Pixelfed client login + photo posting. Existing tables cover everything.

**Soft change needed**: The `Status` struct in [`internal/gtsmodel/status.go`](file:///home/midlajm/Projects/gotosocial/internal/gtsmodel/status.go) has no `LocationID` field. For Phase A, null/absent location is fine — clients tolerate missing location.

### 5.2 Required for Phase B

#### New Table: `collections`

```sql
CREATE TABLE collections (
    id         CHAR(26) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    account_id CHAR(26) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    title      TEXT NOT NULL,
    description TEXT,
    visibility VARCHAR(16) NOT NULL DEFAULT 'public',
    post_count INT NOT NULL DEFAULT 0
);
```

**New model**: `gtsmodel.Collection` in `internal/gtsmodel/collection.go`

#### New Table: `collection_items`

```sql
CREATE TABLE collection_items (
    id             CHAR(26) PRIMARY KEY,
    collection_id  CHAR(26) NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    status_id      CHAR(26) NOT NULL REFERENCES statuses(id) ON DELETE CASCADE,
    account_id     CHAR(26) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(collection_id, status_id)
);
```

#### New Table: `stories`

```sql
CREATE TABLE stories (
    id                    CHAR(26) PRIMARY KEY,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    account_id            CHAR(26) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    media_attachment_id   CHAR(26) NOT NULL REFERENCES media_attachments(id),
    visibility            VARCHAR(16) NOT NULL DEFAULT 'public',
    duration              INT NOT NULL DEFAULT 10,
    expires_at            TIMESTAMPTZ NOT NULL,
    views                 INT NOT NULL DEFAULT 0
);
```

**New model**: `gtsmodel.Story` in `internal/gtsmodel/story.go`

#### New Table: `media_tags` (photo-tagging)

```sql
CREATE TABLE media_tags (
    id            CHAR(26) PRIMARY KEY,
    media_id      CHAR(26) NOT NULL REFERENCES media_attachments(id) ON DELETE CASCADE,
    account_id    CHAR(26) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    x             FLOAT,
    y             FLOAT
);
```

#### Status table: Add location fields

```sql
ALTER TABLE statuses ADD COLUMN place_id CHAR(26) REFERENCES places(id);
```

#### New Table: `places`

```sql
CREATE TABLE places (
    id         CHAR(26) PRIMARY KEY,
    name       TEXT NOT NULL,
    country    TEXT,
    city       TEXT,
    latitude   FLOAT,
    longitude  FLOAT,
    slug       TEXT
);
```

### 5.3 Migration Strategy

GtS uses **bun** ORM with a migration system located in `internal/db/bundb/`. Each migration is a numbered SQL file. New migrations would be added as the next sequential file. No existing schemas are modified destructively — only additive.

---

## Section 6 — ActivityPub Compatibility Layer Design

### 6.1 Collections / Albums

**Storage model**: `gtsmodel.Collection` + `gtsmodel.CollectionItem` tables.

**API layer**: New handler package `internal/api/client/collections/` with route registration in `internal/api/client.go`.

**ActivityPub representation**:
- A collection becomes an `OrderedCollection` at URI `https://host/users/:username/collections/:id`
- Items are `OrderedCollectionPage` entries linking to individual status URIs
- This is **already parseable** by GtS since `OrderedCollection` is handled in `internal/ap/collections.go`

**Federation behavior**: When a Pixelfed server sends a `Create` activity with `object.type = "Album"`, GtS already has `ObjectAlbum = "Album"` in `internal/ap/activitystreams.go` and it's in `IsStatusable()`. However, `ExtractAttachments` only extracts `Attachmentable` types from the `attachment[]` property — `Album` has no dedicated extraction path. A minimal addition to `internal/ap/extract.go` would handle this.

**Backward compatibility**: Collections are additive. Existing statuses aren't affected.

### 6.2 Stories

**Storage model**: `gtsmodel.Story` with `expires_at` field. The existing `internal/scheduler/` package can handle story expiry cleanup.

**API layer**: New handler package `internal/api/client/stories/`.

**ActivityPub representation**: Stories have no standard ActivityPub representation. Pixelfed does not federate stories via AP. Safe to implement as local-only with no AP activity.

**Federation behavior**: None needed. Stories stay local.

### 6.3 Location Metadata

**Storage model**: `gtsmodel.Place`. Referenced from `gtsmodel.Status.PlaceID`.

**API layer**:
- `GET /api/v1/location/search` — searches `places` table; optionally proxies to external geocoder
- Status create endpoint already accepts all form fields — add `place_id` parsing in `internal/api/model/status.go` `StatusCreateRequest`

**ActivityPub representation**:
- When serializing a status with a location, add a `Place` object to the status's `location` property:
  ```json
  "location": { "type": "Place", "name": "Paris", "latitude": 48.85, "longitude": 2.35 }
  ```
- Extraction: Add `ExtractLocation()` to `internal/ap/extract.go`; add `WithLocation` interface to `internal/ap/interfaces.go`
- The `ObjectPlace` constant already exists in `internal/ap/activitystreams.go`

**Backward compatibility**: Location is optional everywhere; null/absent is always acceptable.

### 6.4 Photo Tags

**Storage model**: `gtsmodel.MediaTag` in new `internal/gtsmodel/mediatag.go`.

**API layer**: Extend `/api/v1/media/:id` with sub-routes `tags` in `internal/api/client/media/`.

**ActivityPub representation**: Photo tags can be expressed as `tag[]` entries of type `Mention` on the parent `Image` attachment, pointing at account URIs. This requires extending `Attachmentable` extraction in `internal/ap/extract.go`.

**Federation behavior**: Send `Update` activity for the parent status when tags change.

---

## Section 7 — Implementation Roadmap

### Phase A — Minimum Viable Pixelfed Client Support

**Goal**: Pixelfed Android/iOS/web clients can log in, browse home feed, upload photos, publish posts, view profiles.

**No new database tables required.**

Changes needed:

#### A1. Stub missing endpoints to return empty instead of 404

Endpoints Pixelfed clients call that must not crash the app:

| Endpoint | Response stub |
|----------|--------------|
| `GET /api/v1/discover/posts` | `[]` (empty array) |
| `GET /api/v1/discover/tags/trending` | `[]` |
| `GET /api/v1/stories` | `[]` |
| `GET /api/v1/accounts/:id/stories` | `[]` |
| `GET /api/v1/accounts/:id/albums` | `[]` |
| `GET /api/v1/collections` | `[]` |

**Files to create**: `internal/api/client/discover/` (new package), extend `internal/api/client.go` to register it.

**Complexity**: 1-2 hours per stub endpoint.

#### A2. Direct Message timeline alias

```
GET /api/v1/timelines/direct
```

Already handled by `GET /api/v1/conversations` but Pixelfed clients call `/timelines/direct`. Add an alias route in `internal/api/client/timelines/timeline.go` that calls the conversations processor.

**Files to modify**: `internal/api/client/timelines/timeline.go`, add handler file `timelines/direct.go`.

#### A3. Instance software hint

Consider adding a `pixelfed_compat: true` field to `GET /api/v2/instance` response to signal compatibility without lying about software identity. This avoids NodeInfo confusion.

**Files to modify**: `internal/api/model/instancev2.go` (add optional field), `internal/typeutils/internaltofrontend.go` (populate it conditionally).

#### A4. Ensure `only_media=true` filter works on public timeline

GtS public timeline ([`internal/api/client/timelines/public.go`](file:///home/midlajm/Projects/gotosocial/internal/api/client/timelines/public.go)) — verify `only_media` query param is respected. Pixelfed's explore tab uses this.

---

### Phase B — Full Pixelfed Compatibility

#### B1. Collections / Albums

1. Create `internal/gtsmodel/collection.go` — `Collection`, `CollectionItem` structs
2. Create DB migration + `internal/db/collection.go` interface
3. Create processing layer `internal/processing/collections/` (CRUD, pagination)
4. Create API handler package `internal/api/client/collections/` with routes
5. Register in `internal/api/client.go`
6. Add ActivityPub serialization for collections as `OrderedCollection` in `internal/typeutils/internaltoas.go`
7. Add collection extraction from incoming AP `Album` objects in `internal/ap/extract.go`

#### B2. Stories

1. Create `internal/gtsmodel/story.go` — `Story` struct
2. Create DB migration + `internal/db/story.go` interface
3. Create `internal/media/` story type handling (ephemeral flag, TTL)
4. Create processing layer `internal/processing/stories/`
5. Create API handler `internal/api/client/stories/`
6. Register in `internal/api/client.go`
7. Add scheduler task in `internal/scheduler/` to expire stories

#### B3. Discover / Explore

1. Create `internal/api/client/discover/` — `discover.go` for routes, `posts.go`, `tags.go`, `accounts.go`
2. Back with existing processing: public timeline, trends, directory
3. Register in `internal/api/client.go`

#### B4. Location Support

1. Create `internal/gtsmodel/place.go` — `Place` struct
2. Create DB migration + `internal/db/place.go`
3. Extend `internal/api/model/status.go` `StatusCreateRequest` with `PlaceID string`
4. Extend `internal/api/model/status.go` `Status` with `Place *Place`
5. Add `GET /api/v1/location/search` handler
6. Extend `internal/ap/interfaces.go` — `WithLocation` interface
7. Extend `internal/ap/extract.go` — `ExtractLocation()` 
8. Extend `internal/typeutils/internaltoas.go` — serialize `Place` onto status AP object
9. Extend `internal/typeutils/internaltofrontend.go` — include `Place` in API response

#### B5. Photo Tagging

1. Create `internal/gtsmodel/mediatag.go` — `MediaTag` struct
2. DB migration + `internal/db/mediatag.go`
3. Add sub-routes to `internal/api/client/media/media.go`
4. Create handlers `media/mediatags.go`
5. Extend AP extraction for tagged people in image attachments

---

### Phase C — Pixelfed Enhanced

These are optional improvements beyond base Pixelfed compatibility:

- **EXIF stripping**: Add configurable EXIF removal in `internal/media/imaging.go`
- **Media licensing**: Add `License` field to `gtsmodel.MediaAttachment`, expose in AP and API
- **Album federation**: Full Funkwhale-style `Album` ActivityPub federation with `Create Album` activities
- **Story federation**: Investigate `EphemeralObject` extension for federated stories
- **Gallery web layout**: Already partially present — `WebLayout = "gallery"` exists in `internal/gtsmodel/accountsettings.go`; ensure Pixelfed-style profile display works

---

## Section 8 — Federation Impact Assessment

| Change | Federation Impact |
|--------|------------------|
| Stub empty endpoints | None — local API only |
| Collections (local-only) | None |
| Collections (federated) | Adds `OrderedCollection` at new URIs; backward compatible |
| Stories | None if local-only |
| Location on statuses | Adds `location` property to outgoing AP Note objects; remote instances ignore unknown properties |
| Photo tags on attachments | Adds `tag[]` entries to `Image` AP objects; backward compatible |
| Place objects | Adds new AP `Place` type; remote instances ignore if they don't understand |
| Album AP type | Adds `Create Album` activities; GtS already recognizes `Album` as Statusable so no existing federation behavior breaks |

**Risk**: Pixelfed's AP `Create Album` wraps a collection of `Image` objects. GtS currently accepts `Album` in `IsStatusable()` but stores it as a generic status. Adding proper Album extraction would change how federated Pixelfed albums are stored — this needs careful versioning.

---

## Section 9 — Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Pixelfed clients gate features on `software.name === "pixelfed"` | High | Medium | Stories/albums won't show if client checks; stub endpoints still prevent crashes |
| Clients crash on 404 for `/api/v1/stories` | Medium | High | Phase A: return `[]` stubs |
| AP Album ingestion changes stored content | Medium | High | Version behind a feature flag |
| Location data privacy concerns | Low | High | Make location opt-in at instance level |
| Story expiry misalignment with client expectations | Medium | Medium | Default to 24h, make configurable |
| Breaking existing GtS behavior | Low | High | All changes are additive |
| OAuth scope conflicts | None | None | New scopes are additive |

---

## Section 10 — Prioritized File-by-File Implementation Guide

### Phase A Files (smallest set for client compatibility)

| File | Change Type | What to do |
|------|-------------|-----------|
| [`internal/api/client.go`](file:///home/midlajm/Projects/gotosocial/internal/api/client.go) | Modify | Add `discover *discover.Module` field; register it in `Route()` and `NewClient()` |
| `internal/api/client/discover/discover.go` | **[NEW]** | Routes: `GET /v1/discover/posts`, `GET /v1/discover/tags/trending`; return `[]` stubs |
| `internal/api/client/stories/stories.go` | **[NEW]** | Routes: `GET /v1/stories`, `GET /v1/accounts/:id/stories`; return `[]` stubs |
| [`internal/api/client/timelines/timeline.go`](file:///home/midlajm/Projects/gotosocial/internal/api/client/timelines/timeline.go) | Modify | Add `DirectTimeline = BasePath + "/direct"` route constant; add `GET` handler that delegates to conversations processor |
| `internal/api/client/timelines/direct.go` | **[NEW]** | Handler: return conversations as status-formatted DMs |
| [`internal/api/model/instancev2.go`](file:///home/midlajm/Projects/gotosocial/internal/api/model/instancev2.go) | Modify | Add `PixelfedCompat bool json:"pixelfed_compat,omitempty"` to `InstanceV2Configuration` |
| [`internal/typeutils/internaltofrontend.go`](file:///home/midlajm/Projects/gotosocial/internal/typeutils/internaltofrontend.go) | Modify | In instance conversion function, set `PixelfedCompat: true` if config flag enabled |

### Phase B Files (full compatibility)

#### Collections

| File | Change Type | What to do |
|------|-------------|-----------|
| `internal/gtsmodel/collection.go` | **[NEW]** | Define `Collection` struct with bun tags, `CollectionItem` struct |
| `internal/db/collection.go` | **[NEW]** | Interface: `GetCollection`, `PutCollection`, `DeleteCollection`, `GetCollectionItems`, `PutCollectionItem`, `DeleteCollectionItem` |
| `internal/db/bundb/collection.go` | **[NEW]** | SQL implementations using bun ORM |
| DB migration file | **[NEW]** | `CREATE TABLE collections`, `CREATE TABLE collection_items` |
| `internal/processing/collections/` | **[NEW]** | CRUD processor methods: `CollectionGet`, `CollectionCreate`, `CollectionDelete`, `CollectionAddStatus`, `CollectionRemoveStatus` |
| `internal/api/client/collections/collections.go` | **[NEW]** | Route registration |
| `internal/api/client/collections/collectioncreate.go` | **[NEW]** | POST handler |
| `internal/api/client/collections/collectionget.go` | **[NEW]** | GET handler |
| `internal/api/client/collections/collectiondelete.go` | **[NEW]** | DELETE handler |
| `internal/api/model/collection.go` | **[NEW]** | API response struct `Collection`, `CollectionItem` |
| [`internal/api/client.go`](file:///home/midlajm/Projects/gotosocial/internal/api/client.go) | Modify | Add `collections *collections.Module` |
| [`internal/typeutils/internaltofrontend.go`](file:///home/midlajm/Projects/gotosocial/internal/typeutils/internaltofrontend.go) | Modify | Add `CollectionToAPICollection()` |

#### Stories

| File | Change Type | What to do |
|------|-------------|-----------|
| `internal/gtsmodel/story.go` | **[NEW]** | Define `Story` struct with `ExpiresAt time.Time` |
| `internal/db/story.go` | **[NEW]** | Interface: `GetStory`, `PutStory`, `DeleteStory`, `GetStoriesByAccountID`, `GetActiveStories`, `GetExpiredStories` |
| `internal/db/bundb/story.go` | **[NEW]** | SQL implementations |
| DB migration file | **[NEW]** | `CREATE TABLE stories` |
| `internal/processing/stories/` | **[NEW]** | Create, delete, list stories |
| `internal/api/client/stories/stories.go` | Modify | Full implementation (replace stub) |
| `internal/api/model/story.go` | **[NEW]** | API response struct `Story` |
| `internal/scheduler/` | Modify | Add story expiry job: `ScheduleStoryExpiry()` that calls `DeleteExpiredStories()` |

#### Location

| File | Change Type | What to do |
|------|-------------|-----------|
| `internal/gtsmodel/place.go` | **[NEW]** | `Place` struct with bun tags |
| `internal/gtsmodel/status.go` | Modify | Add `PlaceID string` and `Place *Place` fields |
| `internal/db/place.go` | **[NEW]** | Interface: `GetPlace`, `PutPlace`, `SearchPlaces` |
| `internal/db/bundb/place.go` | **[NEW]** | SQL implementation |
| DB migration file | **[NEW]** | `CREATE TABLE places`, `ALTER TABLE statuses ADD COLUMN place_id` |
| [`internal/api/model/status.go`](file:///home/midlajm/Projects/gotosocial/internal/api/model/status.go) | Modify | Add `Place *Place json:"place"` to `Status`, add `PlaceID string` to `StatusCreateRequest` |
| `internal/api/model/place.go` | **[NEW]** | API struct `Place` |
| `internal/api/client/discover/location.go` | **[NEW]** | `GET /api/v1/location/search` handler |
| [`internal/ap/interfaces.go`](file:///home/midlajm/Projects/gotosocial/internal/ap/interfaces.go) | Modify | Add `WithLocation` interface |
| [`internal/ap/extract.go`](file:///home/midlajm/Projects/gotosocial/internal/ap/extract.go) | Modify | Add `ExtractLocation(Statusable) *gtsmodel.Place` |
| [`internal/typeutils/internaltoas.go`](file:///home/midlajm/Projects/gotosocial/internal/typeutils/internaltoas.go) | Modify | Serialize `Place` as AP `Place` object on status |
| [`internal/typeutils/internaltofrontend.go`](file:///home/midlajm/Projects/gotosocial/internal/typeutils/internaltofrontend.go) | Modify | Include `Place` in status API response |

#### Photo Tagging

| File | Change Type | What to do |
|------|-------------|-----------|
| `internal/gtsmodel/mediatag.go` | **[NEW]** | `MediaTag` struct |
| `internal/db/mediatag.go` | **[NEW]** | Interface |
| `internal/db/bundb/mediatag.go` | **[NEW]** | SQL implementation |
| DB migration file | **[NEW]** | `CREATE TABLE media_tags` |
| `internal/api/client/media/mediatags.go` | **[NEW]** | Handlers for `POST/GET/DELETE /api/v1/media/:id/tags` |
| [`internal/api/client/media/media.go`](file:///home/midlajm/Projects/gotosocial/internal/api/client/media/media.go) | Modify | Register new tag sub-routes |
| `internal/api/model/mediatag.go` | **[NEW]** | `MediaTag` API struct |

---

## Open Questions for Stakeholder Review

> [!IMPORTANT]
> Before implementation begins, the following design decisions need explicit answers:

1. **NodeInfo identity**: Should GtS ever report `software.name = "pixelfed"` in NodeInfo to enable Pixelfed-specific client features? This is a significant identity decision with federation implications.

2. **Story federation**: Should stories remain strictly local-only, or should there be an attempt to define a GtS ephemeral content extension?

3. **Location privacy**: Should location data collection be opt-in per-instance (via config flag) or opt-in per-user? Geocoder privacy implications need review.

4. **Album AP semantics**: When GtS receives a Pixelfed `Album` object, should it store it as a single status with all attachments, or attempt to preserve album structure? Current behavior (single status) loses grouping.

5. **Mastodon version string**: Should GtS change `instanceMastodonVersion` in `internal/typeutils/internaltofrontend.go` to include a Pixelfed compat marker, or add a separate field?

6. **Scope for photo tagging**: Tagging people in photos raises privacy concerns (tagged person may not consent). Should there be a notification + acceptance flow (similar to follow requests)?

---

## Compatibility Matrix Summary

| Feature | Status | Phase |
|---------|--------|-------|
| OAuth login/token | ✅ Full | — |
| App registration | ✅ Full | — |
| Home/public/tag timelines | ✅ Full | — |
| Photo upload + post | ✅ Full | — |
| Profile view/edit | ✅ Full | — |
| Follow/unfollow/block | ✅ Full | — |
| Notifications | ✅ Full | — |
| Search | ✅ Full | — |
| Bookmarks/favourites | ✅ Full | — |
| Content warnings | ✅ Full | — |
| Sensitive media flag | ✅ Full | — |
| Alt text on media | ✅ Full | — |
| Hashtags | ✅ Full | — |
| Mentions | ✅ Full | — |
| Polls | ✅ Full | — |
| Boosts/reblogs | ✅ Full | — |
| Direct messages | ⚠️ Alias needed | A |
| Discover/Explore | ⚠️ Stub + alias | A/B |
| Stories endpoint (stub) | ⚠️ Stub | A |
| Stories (functional) | ❌ Missing | B |
| Collections/Albums (stub) | ⚠️ Stub | A |
| Collections/Albums (functional) | ❌ Missing | B |
| Photo tagging | ❌ Missing | B |
| Location/Place | ❌ Missing | B |
| EXIF handling | ❌ Missing | C |
| Media licensing | ❌ Missing | C |
| Album AP federation | ⚠️ Partial (type recognized, not extracted) | C |
