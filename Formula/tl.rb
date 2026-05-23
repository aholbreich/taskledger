# Homebrew Formula for tl tool (tl)
#
# Install the latest stable release:
#   brew install aholbreich/taskledger/tl
#
# Install from source (HEAD):
#   brew install --HEAD aholbreich/taskledger/tl
#
class Tl < Formula
  desc "Git-native task ledger for human and AI agent coordination"
  homepage "https://github.com/aholbreich/tl"
  license "MIT"
  version "0.4.0"

  # --- Platform-specific binary archives (stable release) ---
  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/aholbreich/tl/releases/download/v#{version}/tl-darwin-amd64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_AMD64_SHA256"
    else
      url "https://github.com/aholbreich/tl/releases/download/v#{version}/tl-darwin-arm64.tar.gz"
      sha256 "REPLACE_WITH_DARWIN_ARM64_SHA256"
    end
  end
  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/aholbreich/tl/releases/download/v#{version}/tl-linux-amd64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_AMD64_SHA256"
    else
      url "https://github.com/aholbreich/tl/releases/download/v#{version}/tl-linux-arm64.tar.gz"
      sha256 "REPLACE_WITH_LINUX_ARM64_SHA256"
    end
  end

  # --- HEAD install (build from source) ---
  head "https://github.com/aholbreich/tl.git", branch: "main"

  depends_on "go" => :build if build.head?

  def install
    if build.head?
      ldflags = "-s -w -X main.version=#{version}"
      system "go", "build", "-o", bin/"tl", "-ldflags", ldflags, "."
    else
      bin.install "tl"
    end
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/tl --version")
  end
end
