/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package version

import "fmt"

type license uint8

const (
	License_MIT license = iota
	License_GNU_GPL_v3
	License_GNU_Affero_GPL_v3
	License_GNU_Lesser_GPL_v3
	License_Mozilla_PL_v2
	License_Apache_v2
	License_Unlicense
	License_Creative_Common_Zero_v1
	License_Creative_Common_Attribution_v4_int
	License_Creative_Common_Attribution_Share_Alike_v4_int
	License_SIL_Open_Font_1_1
)

// nolint: gocritic
func (lic license) GetBoilerPlate(Package, Description, Year, Author string) string {
	switch lic {
	case License_Apache_v2:
		return boiler_Apache2(Year, Author)
	case License_GNU_Affero_GPL_v3:
		return boiler_AGPLv3(Package, Description, Year, Author)
	case License_GNU_GPL_v3:
		return boiler_GPLv3(Package, Description, Year, Author)
	case License_GNU_Lesser_GPL_v3:
		return boiler_LGPLv3(Package, Description, Year, Author)
	case License_MIT:
		return boiler_MIT(Year, Author)
	case License_Mozilla_PL_v2:
		return boiler_MPLv2(Package, Year, Author)
	case License_Unlicense:
		return boiler_Unlicence()
	case License_Creative_Common_Zero_v1:
		return boiler_CC0v1(Year, Author)
	case License_Creative_Common_Attribution_v4_int:
		return boiler_CC_BY_4(Year, Author)
	case License_Creative_Common_Attribution_Share_Alike_v4_int:
		return boiler_CC_SA_4(Year, Author)
	case License_SIL_Open_Font_1_1:
		return boiler_SIL_OFL_11(Year, Author)
	}

	return ""
}

func (lic license) GetLicense() string {
	switch lic {
	case License_Apache_v2:
		return license_apache2()
	case License_GNU_Affero_GPL_v3:
		return license_agpl_v3()
	case License_GNU_GPL_v3:
		return license_gpl_v3()
	case License_GNU_Lesser_GPL_v3:
		return license_lgpl_v3()
	case License_MIT:
		return license_mit()
	case License_Mozilla_PL_v2:
		return license_mozilla_v2()
	case License_Unlicense:
		return boiler_Unlicence()
	case License_Creative_Common_Zero_v1:
		return license_cc0_v1()
	case License_Creative_Common_Attribution_v4_int:
		return license_cc_by_4()
	case License_Creative_Common_Attribution_Share_Alike_v4_int:
		return license_cc_sa_4()
	case License_SIL_Open_Font_1_1:
		return license_sil_ofl_v11()
	}

	return ""
}

func (lic license) GetLicenseName() string {
	switch lic {
	case License_Apache_v2:
		return "Apache License - Version 2.0, January 2004"
	case License_GNU_Affero_GPL_v3:
		return "GNU AFFERO GENERAL PUBLIC LICENSE - Version 3, 19 November 2007"
	case License_GNU_GPL_v3:
		return "GNU GENERAL PUBLIC LICENSE - Version 3, 29 June 2007"
	case License_GNU_Lesser_GPL_v3:
		return "GNU LESSER GENERAL PUBLIC LICENSE - Version 3, 29 June 2007"
	case License_MIT:
		return "MIT License"
	case License_Mozilla_PL_v2:
		return "Mozilla Public License Version 2.0"
	case License_Unlicense:
		return "Free and unencumbered software"
	case License_Creative_Common_Zero_v1:
		return "Creative Commons - CC0 1.0 Universal"
	case License_Creative_Common_Attribution_v4_int:
		return "Creative Commons - Attribution 4.0 International"
	case License_Creative_Common_Attribution_Share_Alike_v4_int:
		return "Creative Commons - Attribution-ShareAlike 4.0 International"
	case License_SIL_Open_Font_1_1:
		return "SIL OPEN FONT LICENSE Version 1.1 - 26 February 2007"
	}

	return ""
}

func boiler_MIT(Year, Author string) string {
	return fmt.Sprintf(`
    MIT License

    Copyright (c) %s %s

    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the "Software"), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is
    furnished to do so, subject to the following conditions:

    The above copyright notice and this permission notice shall be included in all
    copies or substantial portions of the Software.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
    IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
    AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
    LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
    OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
    SOFTWARE.
`, Year, Author)
}

func boiler_AGPLv3(Package, Description, Year, Author string) string {
	return fmt.Sprintf(`
    %s %s
    Copyright (C) %s  %s

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
`, Package, Description, Year, Author)
}

func boiler_GPLv3(Package, Description, Year, Author string) string {
	return fmt.Sprintf(`
    %s %s
    Copyright (C) %s %s

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
`, Package, Description, Year, Author)
}

func boiler_LGPLv3(Package, Description, Year, Author string) string {
	return fmt.Sprintf(`
    %s %s
    Copyright (C) %s  %s

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Lesser General Public License as published
    by the Free Software Foundation, either version 3 of the License.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Lesser General Public License for more details.

    You should have received a copy of the GNU Lesser General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
`, Package, Description, Year, Author)
}

func boiler_MPLv2(Package, Year, Author string) string {
	return fmt.Sprintf(`
    Copyright (C) %s  %s

    This material '%s' is subject to the terms of the Mozilla Public
    License, v. 2.0. If a copy of the MPL was not distributed with this
    file, You can obtain one at http://mozilla.org/MPL/2.0/.
`, Year, Author, Package)
}

func boiler_Apache2(year, author string) string {
	return fmt.Sprintf(`
    Copyright %s %s

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
`, year, author)
}

func boiler_Unlicence() string {
	return `
    This is free and unencumbered software released into the public domain.

    Anyone is free to copy, modify, publish, use, compile, sell, or
    distribute this software, either in source code form or as a compiled
    binary, for any purpose, commercial or non-commercial, and by any
    means.

    In jurisdictions that recognize copyright laws, the author or authors
    of this software dedicate any and all copyright interest in the
    software to the public domain. We make this dedication for the benefit
    of the public at large and to the detriment of our heirs and
    successors. We intend this dedication to be an overt act of
    relinquishment in perpetuity of all present and future rights to this
    software under copyright law.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
    EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
    MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
    IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
    OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
    ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
    OTHER DEALINGS IN THE SOFTWARE.

    For more information, please refer to <http://unlicense.org>
`
}

func boiler_CC0v1(Year, Author string) string {
	return fmt.Sprintf(`
    Copyright %s by %s

    To the extent possible under law, the author(s) have dedicated all 
    copyright and related and neighboring rights to this software to the 
    public domain worldwide. This software is distributed without any warranty.

    You should have received a copy of the CC0 Public Domain Dedication 
    along with this software. If not, see 
        <http://creativecommons.org/publicdomain/zero/1.0/>.
`, Year, Author)
}

func boiler_CC_BY_4(Year, Author string) string {
	return fmt.Sprintf(`
    Copyright %s by %s

    The text of and illustrations in this document are licensed under a 
    Creative Commons Attribution 4.0 International Public License ("CC-BY-4.0").
`, Year, Author)
}

func boiler_CC_SA_4(Year, Author string) string {
	return fmt.Sprintf(`
    Copyright %s by %s

    The text of and illustrations in this document are licensed under a 
    Creative Commons Attribution Share Alike 4.0 International Public License ("CC-SA-4.0").
`, Year, Author)
}

func boiler_SIL_OFL_11(Year, Author string) string {
	return fmt.Sprintf(`
    Copyright (c) %s %s

    This Font Software is licensed under the SIL Open Font License, Version 1.1.
    This license is copied below, and is also available with a FAQ at:
            http://scripts.sil.org/OFL
`, Year, Author)
}
