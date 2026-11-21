import { render, screen } from "@testing-library/react";
import { environmentManager, useQuery } from "@tanstack/react-query";
import { getQueryClient } from "@/lib/query-client";
import { Providers } from "@/app/providers";

function Greeting() {
  const { data, isPending } = useQuery({
    queryKey: ["greeting"],
    queryFn: async () => "hello from react-query",
  });
  return <p>{isPending ? "loading" : data}</p>;
}

describe("getQueryClient", () => {
  afterEach(() => environmentManager.setIsServer(() => typeof window === "undefined"));

  it("reuses one client in the browser", () => {
    expect(getQueryClient()).toBe(getQueryClient());
  });

  it("creates a fresh client for each server request", () => {
    environmentManager.setIsServer(() => true);
    expect(getQueryClient()).not.toBe(getQueryClient());
  });

  it("sets a non-zero default staleTime", () => {
    expect(getQueryClient().getDefaultOptions().queries?.staleTime).toBe(60 * 1000);
  });
});

describe("Providers", () => {
  it("lets useQuery resolve under React 19", async () => {
    render(
      <Providers>
        <Greeting />
      </Providers>,
    );
    expect(await screen.findByText("hello from react-query")).toBeInTheDocument();
  });
});
