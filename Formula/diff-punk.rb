class DiffPunk < Formula
  desc "Local browser viewer for git diff"
  homepage "https://github.com/nathanmsmith/diffpunk"
  head "https://github.com/nathanmsmith/diffpunk.git", branch: "main"

  depends_on "go" => :build
  depends_on "node@24" => :build
  depends_on "git"

  def install
    system "npm", "ci", "--cache=#{buildpath}/.npm"
    system "npm", "run", "build"

    ENV["CGO_ENABLED"] = "0"
    system "go", "build", *std_go_args(ldflags: [
      "-X main.version=#{version}",
      "-X main.development=false",
    ]), "."

    doc.install "README.md", "assets/dist/THIRD_PARTY_LICENSES.txt"
  end

  def caveats
    <<~EOS
      Run diff-punk inside a Git repository with the arguments you would pass to git diff:
        diff-punk
        diff-punk --staged
        diff-punk main...HEAD

      The browser viewer runs until you press Ctrl-C in the terminal.
    EOS
  end
end
