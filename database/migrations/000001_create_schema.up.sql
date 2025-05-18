-- account vote type --
CREATE TYPE vote_type AS ENUM ('up', 'down', 'none');

-- visibility type --
CREATE TYPE visibility AS ENUM ('public', 'private');

-- sort direction type --
CREATE TYPE sort_direction AS ENUM ('asc', 'desc');