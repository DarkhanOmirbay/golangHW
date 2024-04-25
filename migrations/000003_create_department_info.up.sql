CREATE TABLE IF NOT EXISTS department_info(
    id BIGSERIAL PRIMARY KEY,
    department_name VARCHAR(255) NOT NULL,
    staff_quantity INTEGER NOT NULL,
    department_director VARCHAR(255) NOT NULL,
    module_id BIGINT REFERENCES module_info(id)
    );
