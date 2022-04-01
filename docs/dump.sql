CREATE TABLE IF NOT EXISTS public.ulrInfo
(
    request_id text NOT NULL,
    url text NOT NULL,
    app_name text NOT NULL,
    rating NUMERIC (10, 2),
    rating_count INT,
    success boolean NOT NULL,
    last_error text,
    stats json,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (request_id)
);