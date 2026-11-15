/*
 * month_sheet_page.js - Copyright (c) 2023-2026 - Olivier Poncet
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

class MonthSheetPage {
    constructor(document) {
        this.document = document;
        const data = document.querySelector('script[type="application/json"]');
        this.codes = data ? JSON.parse(data.textContent) : {};
    }

    start() {
        for (const row of this.document.querySelectorAll('.month-grid tbody tr')) {
            const whole = row.querySelector('select[name^="day:"]');
            const halves = [...row.querySelectorAll('select[name^="am:"], select[name^="pm:"]')];
            if (whole) {
                whole.addEventListener('change', () => this.fill(whole, halves));
            }
            for (const half of halves) {
                half.addEventListener('change', () => this.edit(whole, half));
            }
        }
    }

    fill(whole, halves) {
        if (whole.value === '') {
            return;
        }
        for (const half of halves) {
            if (this.offers(half, whole.value)) {
                half.value = whole.value;
                this.paint(half);
            }
        }
    }

    offers(select, value) {
        for (const option of select.options) {
            if (option.value === value) {
                return true;
            }
        }
        return false;
    }

    edit(whole, half) {
        if (whole) {
            whole.value = '';
        }
        this.paint(half);
    }

    paint(select) {
        const cell = select.closest('td');
        for (const name of [...cell.classList]) {
            if (name.startsWith('day-') || name.startsWith('code-')) {
                cell.classList.remove(name);
            }
        }
        cell.classList.add(this.codes[select.value] || cell.dataset.blank || 'day-empty');
    }
}

new MonthSheetPage(document).start();
