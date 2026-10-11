-- ============================================================
-- StudHub Club Service
-- Initial Database Schema
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE club_status AS ENUM (
    'ACTIVE',
    'INACTIVE',
    'ARCHIVED'
);

CREATE TYPE club_visibility AS ENUM (
    'PUBLIC',
    'PRIVATE'
);

CREATE TYPE club_join_type AS ENUM (
    'OPEN',
    'APPROVAL_REQUIRED',
    'INVITE_ONLY'
);

CREATE TYPE club_member_role AS ENUM (
    'OWNER',
    'ADMIN',
    'MODERATOR',
    'MEMBER'
);

CREATE TYPE club_member_status AS ENUM (
    'ACTIVE',
    'BANNED',
    'LEFT'
);

CREATE TYPE join_request_status AS ENUM (
    'PENDING',
    'APPROVED',
    'REJECTED',
    'CANCELLED'
);


-- ============================================================
-- CLUB CATEGORIES
-- ============================================================

CREATE TABLE club_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(100) NOT NULL,
    description TEXT,

    icon_url TEXT,

    display_order INTEGER NOT NULL DEFAULT 0,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_club_categories_name
        UNIQUE (name),

    CONSTRAINT chk_club_categories_display_order
        CHECK (display_order >= 0)
);


-- ============================================================
-- CLUBS
-- ============================================================

CREATE TABLE clubs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(150) NOT NULL,

    slug VARCHAR(180) NOT NULL,

    description TEXT,

    category_id UUID NOT NULL,

    -- User who created the club
    created_by UUID NOT NULL,

    -- College associated with the club.
    -- Kept as UUID because College Service owns college data.
    college_id UUID,

    logo_url TEXT,

    banner_url TEXT,

    status club_status NOT NULL DEFAULT 'ACTIVE',

    visibility club_visibility NOT NULL DEFAULT 'PUBLIC',

    join_type club_join_type NOT NULL DEFAULT 'OPEN',

    -- Used to control how clubs are displayed.
    display_order INTEGER NOT NULL DEFAULT 0,

    -- Cached count for quick listing.
    -- Actual membership is maintained in club_members.
    member_count INTEGER NOT NULL DEFAULT 0,

    -- ID of the discussion channel created by Chat Service.
    channel_id UUID,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMPTZ,

    CONSTRAINT fk_clubs_category
        FOREIGN KEY (category_id)
        REFERENCES club_categories(id),

    CONSTRAINT uq_clubs_slug
        UNIQUE (slug),

    CONSTRAINT chk_clubs_display_order
        CHECK (display_order >= 0),

    CONSTRAINT chk_clubs_member_count
        CHECK (member_count >= 0)
);


-- ============================================================
-- CLUB MEMBERS
-- ============================================================

CREATE TABLE club_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    club_id UUID NOT NULL,

    -- User Service owns the actual user record.
    user_id UUID NOT NULL,

    role club_member_role NOT NULL DEFAULT 'MEMBER',

    status club_member_status NOT NULL DEFAULT 'ACTIVE',

    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    left_at TIMESTAMPTZ,

    banned_at TIMESTAMPTZ,

    CONSTRAINT fk_club_members_club
        FOREIGN KEY (club_id)
        REFERENCES clubs(id)
        ON DELETE CASCADE,

    -- A user can have only one membership record per club.
    CONSTRAINT uq_club_members_club_user
        UNIQUE (club_id, user_id)
);


-- ============================================================
-- CLUB JOIN REQUESTS
-- ============================================================

CREATE TABLE club_join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    club_id UUID NOT NULL,

    user_id UUID NOT NULL,

    status join_request_status NOT NULL DEFAULT 'PENDING',

    -- Optional message from the student while requesting to join.
    message TEXT,

    -- User/admin who processed the request.
    reviewed_by UUID,

    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    reviewed_at TIMESTAMPTZ,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_join_requests_club
        FOREIGN KEY (club_id)
        REFERENCES clubs(id)
        ON DELETE CASCADE
);


-- ============================================================
-- INDEXES
-- ============================================================

-- Category filtering
CREATE INDEX idx_clubs_category_id
    ON clubs(category_id);

-- College filtering
CREATE INDEX idx_clubs_college_id
    ON clubs(college_id);

-- Active club listing
CREATE INDEX idx_clubs_status
    ON clubs(status);

-- Public/private filtering
CREATE INDEX idx_clubs_visibility
    ON clubs(visibility);

-- Sorting
CREATE INDEX idx_clubs_display_order
    ON clubs(display_order);

CREATE INDEX idx_clubs_created_at
    ON clubs(created_at DESC);

CREATE INDEX idx_clubs_member_count
    ON clubs(member_count DESC);


-- Club member lookups
CREATE INDEX idx_club_members_club_id
    ON club_members(club_id);

CREATE INDEX idx_club_members_user_id
    ON club_members(user_id);

CREATE INDEX idx_club_members_status
    ON club_members(status);


-- Join request lookups
CREATE INDEX idx_join_requests_club_id
    ON club_join_requests(club_id);

CREATE INDEX idx_join_requests_user_id
    ON club_join_requests(user_id);

CREATE INDEX idx_join_requests_status
    ON club_join_requests(status);


-- ============================================================
-- UNIQUE PENDING JOIN REQUEST
-- ============================================================

CREATE UNIQUE INDEX uq_pending_join_request
    ON club_join_requests(club_id, user_id)
    WHERE status = 'PENDING';