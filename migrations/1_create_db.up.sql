CREATE TABLE "promotion" (
    "id"             serial PRIMARY KEY,
    "description"    varchar,
    "percentage"     decimal,
    "start_date"     timestamp NOT NULL,
	"end_date"       timestamp NOT NULL,
    "created_at"     timestamp default now(),
    "updated_at"     timestamp default now(),
    "deleted_at"     timestamp
);

CREATE TABLE "billing" (
    "id"                serial PRIMARY KEY,
    "promotion_id"      integer,
    "total"	            decimal NOT NULL,
    "created_at"        timestamp default now(),
    "updated_at"        timestamp default now(),
    "deleted_at"        timestamp
);

CREATE TABLE "billing_detail" (
    "id"             serial PRIMARY KEY,
    "billing_id"     integer NOT NULL,
    "medicine_id"    integer NOT NULL,
    "medicine_name"  varchar NOT NULL,
    "medicine_price" decimal NOT NULL,
    "created_at"     timestamp default now(),
    "updated_at"     timestamp default now(),
    "deleted_at"     timestamp
);

CREATE TABLE "medicine" (
    "id"             serial PRIMARY KEY,
    "name"           varchar NOT NULL,
    "price"	         decimal NOT NULL,
    "location"	     varchar,
    "created_at"     timestamp default now(),
    "updated_at"     timestamp default now(),
    "deleted_at"     timestamp
);

ALTER TABLE "billing"
    ADD FOREIGN KEY ("promotion_id") REFERENCES "promotion" ("id");

ALTER TABLE "billing_detail"
    ADD FOREIGN KEY ("billing_id") REFERENCES "billing" ("id");

ALTER TABLE "billing_detail"
    ADD FOREIGN KEY ("medicine_id") REFERENCES "medicine" ("id");

