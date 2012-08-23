require 'formula'

class Visor < Formula
  homepage 'http://github.com/soundcloud/visor'
  url 'https://github.com/soundcloud/visor/zipball/master'
  depends_on 'go'
  skip_clean 'bin'
  version '0.5.6'


  def install
    begin
      system("which hg")
    rescue
      system "brew install hg"
    end

    begin
      system("hg --version")
    rescue
      hg_path = `which hg`
      puts <<-EOF

Your mercurial installation is broken. Please delete your current mercurial installation
and reinstall with brew by executing:

  sudo mv #{hg_path.chomp} /tmp
  brew install --force mercurial || brew update --force mercurial



      EOF

      exit 1
    end
    ENV['GOPATH'] = buildpath
    ENV['GOBIN'] = "#{prefix}/bin"
    system "make"
  end

  def test
    system "visor --version"
  end
end
