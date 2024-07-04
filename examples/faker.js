// source: https://github.com/szkiba/xk6-faker/blob/v0.3.0/examples/custom-faker.js
import { Faker } from "k6/x/faker";

const faker = new Faker(11);

export default function () {
  console.log(faker.person.firstName());
}

// output: Josiah
