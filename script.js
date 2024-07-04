import { Faker } from "k6/x/faker";
import sql from "k6/x/sql";

const db = sql.open("sqlite3", "./test-users.db");

export function setup() {
  db.exec(`
  CREATE TABLE IF NOT EXISTS users (
    sub varchar PRIMARY KEY,
    name varchar NOT NULL,
    email varchar NOT NULL
  );`);

  const faker = new Faker(11);

  db.exec(`
    INSERT OR REPLACE INTO users (sub, name, email) VALUES (
      '${faker.internet.username()}',
      '${faker.person.firstName()} ${faker.person.lastName()}',
      '${faker.person.email()}'
    );`);
}

export function teardown() {
  db.close();
}

export default function () {
  const results = sql.query(db, "SELECT * FROM users");

  for (const row of results) {
    const { sub, name, email } = row;

    console.log({ sub, name, email });
  }
}
