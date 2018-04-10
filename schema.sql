CREATE TABLE ticker (
  pair        varchar(10),
  high        double precision,
  low         double precision,
  avg         double precision,
  vol         double precision,
  vol_cur     double precision,
  last        double precision,
  buy         double precision,
  sell        double precision,
  updated     bigint,
  PRIMARY KEY (pair, updated)
);

CREATE INDEX ticker_updated ON ticker (updated);
