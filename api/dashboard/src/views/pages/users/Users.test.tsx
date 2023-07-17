import React from "react";

import { act, render, screen } from "@testing-library/react";
import { Mock, vi } from "vitest";

import { getUserList } from "../../../Controllers/UserControllers/UsersApi";

import Users from "./Users";

vi.mock("../../../Controllers/UserControllers/UsersApi", () => {
  return { getUserList: vi.fn() };
});

const flushPromises = (): Promise<void> => new Promise(setImmediate);

vi.mock("react-router", () => {
  const useNavigate = vi.fn(() => vi.fn());
  return { useNavigate };
});

describe("User Page test", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("shows message about no user entries", async () => {
    (getUserList as Mock).mockRejectedValueOnce({
      response: { data: { error: { message: "" } } },
    });

    render(<Users />);
    expect(getUserList).toBeCalledWith({
      limit: 10,
      offset: 0,
      orderBy: "id asc",
    });

    await act(flushPromises);

    const linkElement = screen.getByText("No items found");
    expect(linkElement).not.toBeNull();
  });

  it("shows users in table", async () => {
    (getUserList as Mock).mockResolvedValueOnce({
      data: [
        { email: "mrBombastic", verified: false, name: "tell me Fantastic" },
      ],
    });
    render(<Users />);
    expect(getUserList).toBeCalledWith({
      limit: 10,
      offset: 0,
      orderBy: "id asc",
    });

    await act(flushPromises);

    const linkElement = screen.queryByText("mrBombastic");
    expect(linkElement).not.toBeNull();
  });
});
