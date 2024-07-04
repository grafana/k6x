"use k6 >= 0.50";
"use k6 with k6/x/faker >= 0.3.0";
"use k6 with k6/x/sql >= 0.4.0";

import faker from "./faker.js";
import sqlite from "./sqlite.js";

export { setup, teardown } from "./sqlite.js";

export default () => {
  faker();
  sqlite();
};
