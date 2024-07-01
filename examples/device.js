import { parse } from "k6/x/yaml";

const source = `
- name: foo
- name: bar
`;

const devices = parse(source);

export function getDevice() {
  return devices[0];
}
