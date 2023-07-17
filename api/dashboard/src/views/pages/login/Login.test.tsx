import React from "react";

import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Mock, expect, it, vi } from "vitest";

import { userLoginInfo } from "../../../Controllers/OAuth/OAuth";

import Login from "./Login";

vi.mock("../../../Controllers/OAuth/OAuth", () => {
  return { userLoginInfo: vi.fn() };
});

const flushPromises = (): Promise<void> => new Promise(setImmediate);

vi.mock("react-router", () => {
  const useNavigate = vi.fn(() => vi.fn());
  return { useNavigate };
});

describe("Login test list", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("shows an error if login fails", async () => {
    const user = userEvent.setup();
    (userLoginInfo as Mock).mockRejectedValueOnce({
      response: {
        data: {
          error_description: "API Error",
        },
      },
    });
    let isLoggedIn = false;
    render(
      <Login
        onLoggedChange={(_isLoggedIn) => {
          isLoggedIn = _isLoggedIn;
        }}
      />
    );
    const usernameInput = screen.getByPlaceholderText("example@email.com");
    const passwordInput = screen.getByPlaceholderText("Password");
    const loginButton = screen.getByText("Continue");
    await user.click(usernameInput);
    await user.keyboard("asdfasdasdf");
    await user.click(passwordInput);
    await user.keyboard("12341234");
    await user.click(loginButton);

    expect(userLoginInfo).toBeCalledWith({
      username: "asdfasdasdf",
      password: "12341234",
    });

    await act(flushPromises);

    const linkElement = screen.getByText("API Error");
    expect(linkElement).not.toBeNull();
    expect(isLoggedIn).toBe(false);
  });

  it("saves access token to local storage", async () => {
    const user = userEvent.setup();
    (userLoginInfo as Mock).mockResolvedValueOnce({ status: 200, data: {} });
    let isLoggedIn = false;
    render(
      <Login
        onLoggedChange={(_isLoggedIn) => {
          isLoggedIn = _isLoggedIn;
        }}
      />
    );
    const usernameInput = screen.getByPlaceholderText("example@email.com");
    const passwordInput = screen.getByPlaceholderText("Password");
    const loginButton = screen.getByText("Continue");
    await user.click(usernameInput);
    await user.keyboard("asdfasdasdf");
    await user.click(passwordInput);
    await user.keyboard("12341234");
    await user.click(loginButton);

    expect(userLoginInfo).toBeCalledWith({
      username: "asdfasdasdf",
      password: "12341234",
    });

    await act(flushPromises);

    const linkElement = screen.queryByText(
      "Username or Password are incorrect."
    );
    expect(linkElement).toBeNull();
    expect(isLoggedIn).toBe(true);
  });
});
