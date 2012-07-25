require 'formula'

class Visor < Formula
  homepage 'http://github.com/soundcloud/visor'
  url 'https://github.com/soundcloud/visor/zipball/master'
  #md5 'd01e901a9bd781d0104990b2ece77bf0'
  depends_on 'go'
  version '0.5.0'
  skip_clean 'bin'


  def install
    system "make gobuild"
    system "make DESTDIR=#{prefix} install"
  end

  def test
    # This test will fail and we won't accept that! It's enough to just replace
    # "false" with the main program this formula installs, but it'd be nice if you
    # were more thorough. Run the test with `brew test visor`.
    system "visor --version"
  end
end
