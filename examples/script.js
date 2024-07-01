"use k6 >= 0.49";
"use k6 with k6/x/faker >= 0.2.0";

import { newUser } from "./user.js";
import { getDevice } from "./device.js";

export default () => {
  const user = newUser("John");
  console.log(user);
  console.log(getDevice());
};
