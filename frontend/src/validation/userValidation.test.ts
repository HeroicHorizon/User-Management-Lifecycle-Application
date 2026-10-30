import { describe, expect, it } from "vitest";
import { validateUser } from "./userValidation";

const valid = {
  firstName: "John",
  lastName: "Doe",
  email: "john@example.com",
  phone: "+911234567890",
  department: "Eng",
  designation: "Dev",
};

describe("validateUser", () => {
  it("accepts valid input and an empty phone", () => {
    expect(validateUser(valid)).toEqual({});
    expect(validateUser({ ...valid, phone: "" })).toEqual({});
  });
  it("flags whitespace-only required fields", () => {
    expect(validateUser({ ...valid, firstName: "   " }).firstName).toBeDefined();
  });
  it("flags invalid email and phone", () => {
    const e = validateUser({ ...valid, email: "nope", phone: "abc" });
    expect(e.email).toBeDefined();
    expect(e.phone).toBeDefined();
  });
});
