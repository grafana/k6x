import "k6/x/faker";

class UserAccount {
  constructor(name) {
    this.name = name;
    this.id = Math.floor(Math.random() * Number.MAX_SAFE_INTEGER);
  }
}

function newUser(name) {
  return new UserAccount(name);
}

export { newUser };
