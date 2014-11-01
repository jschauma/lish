%define name		lish
%define release		1
%define version 	0.2
%define mybuilddir	${HOME}/redhat/BUILD/%{name}-%{version}-root

BuildRoot:		%{mybuilddir}
Summary:		limited shell
License: 		BSD
Name: 			%{name}
Version: 		%{version}
Release: 		%{release}
Source: 		%{name}-%{version}.tar.gz
Prefix: 		/usr
Group: 			Development/Tools

%description
'lish' is a very simple, restricted command-line interpreter or shell.
Its goal is to allow execution of only a fixed set of commands, either
interactively or read from standard in.  'lish' is suitable to be used as
the ssh(8) ForceCommand or via the 'command=' restriction in an ssh public
key.

%prep
%setup -q

%build

%install
mkdir -p %{mybuilddir}/usr/bin
mkdir -p %{mybuilddir}/usr/share/man/man1
install -c -m 755 src/%{name} %{mybuilddir}/usr/bin/%{name}
install -c -m 444 doc/%{name}.1 %{mybuilddir}/usr/share/man/man1/%{name}.1

%files
%defattr(0444,root,root)
%attr(0755,root,root) /usr/bin/%{name}
%doc /usr/share/man/man1/%{name}.1.gz

%changelog
* Tue Oct 28 2014 - jschauma@twitter.com
- 0.2
-  ensure SSH_ORIGINAL_COMMAND is only used if set and doesn't override
   any otherwise given '-c' commands
-  gofmt via drobinson
-  log current working directory when executing commands

* Tue Oct 28 2014 - jschauma@twitter.com
- 0.1
-  Hello World!
