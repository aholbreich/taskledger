# Homebrew Formula for taskledger (tl)
#
# Install from source (HEAD):
#   brew install --HEAD ./Formula/taskledger.rb
#
# After the first GitHub Release, stable installs work too:
#   brew install taskledger         # from a tap
#
class Taskledger < Formula
  desc "Git-native task ledger for human and AI agent coordination"
  homepage "https://github.com/aholbreich/taskledger"
  license "MIT"

  # --- HEAD install (build from source) ---
  head "https://github.com/aholbreich/taskledger.git", branch: "main"

  # --- Stable release (populate after first release) ---
  # url "https://github.com/aholbreich/taskledger/archive/refs/tags/v0.4.0.tar.gz"
  # sha256 "REPLACE_WITH_ACTUAL_SHA256"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X main.version=#{version}"
    system "go", "build", "-o", bin/"tl", "-ldflags", ldflags, "."
  end

  test do
    # --version exits 0 and prints a version line.
    assert_match version.to_s, shell_output("#{bin}/tl --version")
  end
end
