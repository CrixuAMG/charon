class Charon < Formula
  desc "Terminal project navigator and launcher for Kitty"
  homepage "https://github.com/crixuamg/charon"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/crixuamg/charon/releases/download/v#{version}/charon-darwin-arm64.tar.gz"
      # sha256 will be filled in during release
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
    on_intel do
      url "https://github.com/crixuamg/charon/releases/download/v#{version}/charon-darwin-amd64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/crixuamg/charon/releases/download/v#{version}/charon-linux-arm64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
    on_intel do
      url "https://github.com/crixuamg/charon/releases/download/v#{version}/charon-linux-amd64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
  end

  def install
    bin.install "charon"
  end

  test do
    system "#{bin}/charon", "help"
  end
end
