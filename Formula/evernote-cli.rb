class EvernoteCli < Formula
  desc "CLI tool for interacting with Evernote"
  homepage "https://github.com/cloudmanic/evernote-cli"
  license "MIT"
  version "latest"

  # Download pre-built binaries for macOS from GitHub releases
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/cloudmanic/evernote-cli/releases/latest/download/evernote-cli-darwin-amd64.tar.gz"
    # SHA256 will be updated automatically by Homebrew when using 'latest'
  elsif OS.mac? && Hardware::CPU.arm?
    url "https://github.com/cloudmanic/evernote-cli/releases/latest/download/evernote-cli-darwin-arm64.tar.gz"
    # SHA256 will be updated automatically by Homebrew when using 'latest'
  end

  def install
    if Hardware::CPU.intel?
      bin.install "evernote-cli-darwin-amd64" => "evernote-cli"
    elsif Hardware::CPU.arm?
      bin.install "evernote-cli-darwin-arm64" => "evernote-cli"
    end
  end

  test do
    # Test that the binary runs and shows help
    output = shell_output("#{bin}/evernote-cli --help")
    assert_match "A CLI tool to interact with Evernote", output
    assert_match "Available Commands:", output
    assert_match "init", output
    assert_match "search", output
  end
end