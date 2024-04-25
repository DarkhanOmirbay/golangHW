CREATE TABLE IF NOT EXISTS tokens (
                                      hash bytea PRIMARY KEY,
                                      user_id bigint NOT NULL REFERENCES user_info ON DELETE CASCADE,
                                      expiry timestamp(0) with time zone NOT NULL,
                                                              scope text NOT NULL
                                                              );
