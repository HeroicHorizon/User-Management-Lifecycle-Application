export type UserStatus = "ACTIVE" | "INACTIVE";

export interface User {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  department: string;
  designation: string;
  status: UserStatus;
  createdAt: string;
  updatedAt: string;
}

export interface UserInput {
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  department: string;
  designation: string;
}

export interface UserList {
  items: User[];
  page: number;
  pageSize: number;
  total: number;
}

export interface UserStats {
  total: number;
  active: number;
  inactive: number;
}

export interface ListParams {
  page: number;
  pageSize: number;
  search: string;
  status: UserStatus | "";
}
