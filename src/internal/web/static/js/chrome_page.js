/*
 * chrome_page.js - Copyright (c) 2023-2026 - Olivier Poncet
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

class ChromePage {
    constructor(document) {
        this.document = document;
        this.window = document.defaultView;
        this.root = document.documentElement;
        this.scheme = document.querySelector('meta[name="color-scheme"]');
        this.toggle = document.querySelector('button.theme-toggle');
        this.icon = (this.toggle === null) ? null : this.toggle.querySelector('img');
        this.menus = [...document.querySelectorAll('details.nav-menu, details.user-menu, details.account-delete')];
        this.theme = '';
    }

    start() {
        this.theme = this.stored();
        this.apply(this.theme);
        if (this.toggle !== null) {
            this.toggle.hidden = false;
            this.toggle.addEventListener('click', () => this.switchTheme());
        }
        this.document.addEventListener('click', (event) => this.closeOutside(event.target));
        this.document.addEventListener('keydown', (event) => this.closeOnEscape(event));
        for (const menu of this.menus) {
            menu.addEventListener('toggle', () => this.focusField(menu));
        }
    }

    focusField(menu) {
        const field = menu.querySelector('form input[type="text"]');
        if (menu.open && (field !== null)) {
            field.focus();
        }
    }

    stored() {
        try {
            const theme = this.window.localStorage.getItem('theme');
            return ((theme === 'light') || (theme === 'dark')) ? theme : '';
        } catch {
            return '';
        }
    }

    store(theme) {
        try {
            this.window.localStorage.setItem('theme', theme);
        } catch {
            return;
        }
    }

    effective() {
        if (this.theme !== '') {
            return this.theme;
        }
        return this.window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }

    switchTheme() {
        this.theme = (this.effective() === 'dark') ? 'light' : 'dark';
        this.store(this.theme);
        this.apply(this.theme);
    }

    apply(theme) {
        if (theme !== '') {
            this.root.dataset.theme = theme;
            this.scheme.content = theme;
        }
        if (this.toggle !== null) {
            const dark = this.effective() === 'dark';
            this.toggle.setAttribute('aria-pressed', String(dark));
            this.icon.src = dark ? '/static/img/icons/moon.svg' : '/static/img/icons/sun.svg';
        }
    }

    closeOutside(target) {
        for (const menu of this.menus) {
            if (menu.open && !menu.contains(target)) {
                menu.open = false;
            }
        }
    }

    closeOnEscape(event) {
        if (event.key !== 'Escape') {
            return;
        }
        for (const menu of this.menus) {
            if (menu.open) {
                menu.open = false;
                menu.querySelector('summary').focus();
            }
        }
    }
}

new ChromePage(document).start();
