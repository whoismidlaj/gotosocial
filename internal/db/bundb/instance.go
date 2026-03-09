// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package bundb

import (
	"context"
	"errors"
	"math"
	"slices"
	"sync"
	"time"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type instanceDB struct {
	db    *bun.DB
	state *state.State

	// Used to lock when
	// adding instance errors.
	sync.Mutex
}

func (i *instanceDB) CountInstanceAccounts(ctx context.Context) (int, error) {
	// Check for a cached instance accounts count. If present return this.
	if n := i.state.Caches.DB.LocalInstance.Accounts.Load(); n != nil {
		return *n, nil
	}

	count, err := i.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		// Just select IDs.
		Column("account.id").
		// Local accounts only.
		Where("? IS NULL", bun.Ident("account.domain")).
		// Ignore instance account.
		Where("? != ?", bun.Ident("account.username"), config.GetHost()).
		// Exclude suspended accounts.
		Where("? IS NULL", bun.Ident("account.suspended_at")).
		Count(ctx)
	if err != nil {
		return 0, err
	}

	// Update cached instance accounts count value.
	i.state.Caches.DB.LocalInstance.Accounts.Store(&count)
	return count, nil
}

func (i *instanceDB) CountInstanceStatuses(ctx context.Context) (int, error) {
	// Check for a cached instance statuses count. If present return this.
	if n := i.state.Caches.DB.LocalInstance.Statuses.Load(); n != nil {
		return *n, nil
	}

	// Select from local count view.
	var count int
	if err := i.db.
		NewSelect().
		Table("statuses_local_count_view").
		Scan(ctx, &count); err != nil {
		return 0, err
	}

	// Update cached instance statuses count value.
	i.state.Caches.DB.LocalInstance.Statuses.Store(&count)
	return count, nil
}

func (i *instanceDB) CountInstancePeers(ctx context.Context) (int, error) {
	// Check for a cached instance peers count. If present return this.
	if n := i.state.Caches.DB.LocalInstance.Peers.Load(); n != nil {
		return *n, nil
	}

	// Select just the domain
	// part of all known instances.
	domains := []string{}
	if err := i.db.
		NewSelect().
		Table("instances").
		Column("domain").
		Scan(ctx, &domains); err != nil {
		return 0, err
	}

	var count int
	for _, domain := range domains {
		// For each domain, check if
		// we're federating with it.
		blocked, err := i.state.DB.IsDomainBlocked(ctx, domain)
		if err != nil {
			return 0, gtserror.Newf("db error checking block: %w", err)
		}

		if blocked {
			// Doesn't count as a
			// peer if it's blocked.
			continue
		}

		// Count as a peer.
		count++
	}

	// Update cached instance peers count value.
	i.state.Caches.DB.LocalInstance.Peers.Store(&count)
	return count, nil
}

func (i *instanceDB) GetInstancePeers(ctx context.Context, includeSuspended bool) ([]*gtsmodel.Instance, error) {
	// Select just the domain
	// part of all known instances.
	domains := []string{}
	if err := i.db.
		NewSelect().
		Table("instances").
		Column("domain").
		Scan(ctx, &domains); err != nil {
		return nil, err
	}

	if len(domains) == 0 {
		// Empty response.
		return make([]*gtsmodel.Instance, 0), nil
	}

	instances := make([]*gtsmodel.Instance, 0, len(domains))
	for _, domain := range domains {
		if !includeSuspended {
			// Ensure peer not blocked.
			blocked, err := i.state.DB.IsDomainBlocked(ctx, domain)
			if err != nil {
				return nil, gtserror.Newf("db error checking block: %w", err)
			}

			if blocked {
				// Skip this one.
				continue
			}
		}

		// Select instance.
		instance, err := i.GetInstance(ctx, domain)
		if err != nil {
			log.Errorf(ctx, "db error getting instance %q: %v", domain, err)
			continue
		}

		// Append to return slice.
		instances = append(instances, instance)
	}

	return instances, nil
}

func (i *instanceDB) GetInstance(ctx context.Context, domain string) (*gtsmodel.Instance, error) {
	// Normalize the domain as punycode.
	var err error
	domain, err = util.Punify(domain)
	if err != nil {
		return nil, gtserror.Newf("error punifying domain %s: %w", domain, err)
	}

	return i.getInstance(
		"Domain",
		func(instance *gtsmodel.Instance) error {
			return i.db.NewSelect().
				Model(instance).
				Where("? = ?", bun.Ident("instance.domain"), domain).
				Scan(ctx)
		},
		domain,
	)
}

func (i *instanceDB) GetInstanceByID(ctx context.Context, id string) (*gtsmodel.Instance, error) {
	return i.getInstance(
		"ID",
		func(instance *gtsmodel.Instance) error {
			return i.db.NewSelect().
				Model(instance).
				Where("? = ?", bun.Ident("instance.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (i *instanceDB) GetInstancesPage(
	ctx context.Context,
	page *paging.Page,
	domain string,
	orderBy gtsmodel.InstanceOrderBy,
	undeliverableOnly bool,
) ([]*gtsmodel.Instance, error) {
	var (
		// Extract page params.
		minID = page.Min.Value
		maxID = page.Max.Value
		limit = page.Limit
		order = page.Order()

		// Pre-allocate slice of IDs.
		instanceIDs = make([]string, 0, limit)

		// We know orderBy is either Latest or Alphabetical.
		orderByAlphabetical = orderBy == gtsmodel.InstanceOrderByAlphabetical
	)

	q := i.db.
		NewSelect().
		// Select just the ID
		// of each instance.
		Column("instance.id").
		TableExpr("? AS ?", bun.Ident("instances"), bun.Ident("instance"))

	if undeliverableOnly {
		q = q.Where("? IS NOT NULL", bun.Ident("instance.delivery_errors_count"))
	}

	if domain != "" {
		// Normalize the
		// domain as punycode.
		var err error
		domain, err = util.Punify(domain)
		if err != nil {
			return nil, gtserror.Newf("error punifying domain %s: %w", domain, err)
		}

		// Get any instances *starting with* the given domain.
		q = q.Where("? LIKE ?", bun.Ident("instance.domain"), domain+"%")
	}

	// Paging parameters.

	if maxID != "" {
		if orderByAlphabetical {
			// Get instance for max ID.
			maxIDInstance, err := i.GetInstanceByID(
				gtscontext.SetBarebones(ctx),
				maxID,
			)
			if err != nil {
				err := gtserror.Newf("db error getting maxID instance %s: %w", maxID, err)
				return nil, err
			}
			// Order alpahetically (a-z) by domain.
			q = q.Where("? > ?", bun.Ident("instance.domain"), maxIDInstance.Domain)
		} else {
			// Order by ID, which indicates creation time.
			q = q.Where("? < ?", bun.Ident("instance.id"), maxID)
		}
	}

	if minID != "" {
		if orderByAlphabetical {
			// Get instance for min ID.
			minIDInstance, err := i.GetInstanceByID(
				gtscontext.SetBarebones(ctx),
				minID,
			)
			if err != nil {
				err := gtserror.Newf("db error getting minID instance %s: %w", minID, err)
				return nil, err
			}
			// Order alpahetically (a-z) by domain.
			q = q.Where("? < ?", bun.Ident("instance.domain"), minIDInstance.Domain)
		} else {
			// Order by ID, which indicates creation time.
			q = q.Where("? > ?", bun.Ident("instance.id"), minID)
		}
	}

	switch {
	case !orderByAlphabetical && order == paging.OrderDescending:
		// Order by ID (paging down).
		q = q.Order("instance.id DESC")

	case !orderByAlphabetical && order == paging.OrderAscending:
		// Order by ID (paging up).
		q = q.Order("instance.id ASC")

	case orderByAlphabetical && order == paging.OrderDescending:
		// Order alphabetically (paging down).
		// Z > A in ASCII so use ASC.
		q = q.Order("instance.domain ASC")

	case orderByAlphabetical && order == paging.OrderAscending:
		// Order alphabetically (paging up).
		// A < Z in ASCII so use DESC.
		q = q.Order("instance.domain DESC")
	}

	// Limit amount of
	// instances returned.
	q = q.Limit(limit)

	// Run the query.
	if err := q.Scan(ctx, &instanceIDs); err != nil {
		return nil, err
	}

	count := len(instanceIDs)
	if count == 0 {
		// Nothing for
		// this query.
		return nil, nil
	}

	// Preallocate slice of instances,
	// and fetch each instance by ID.
	instances := make([]*gtsmodel.Instance, 0, count)
	for _, instanceID := range instanceIDs {
		instance, err := i.GetInstanceByID(ctx, instanceID)
		if err != nil {
			err := gtserror.Newf("db error getting instance: %w", err)
			return nil, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func (i *instanceDB) getInstance(
	lookup string,
	dbQuery func(*gtsmodel.Instance) error, keyParts ...any,
) (*gtsmodel.Instance, error) {
	// Fetch instance from database cache with loader callback
	return i.state.Caches.DB.Instance.LoadOne(lookup, func() (*gtsmodel.Instance, error) {
		// Not cached! Perform database query.
		var instance gtsmodel.Instance
		if err := dbQuery(&instance); err != nil {
			return nil, err
		}

		return &instance, nil
	}, keyParts...)
}

func (i *instanceDB) PutInstance(ctx context.Context, instance *gtsmodel.Instance) error {
	// Normalize the domain as punycode, note the extra
	// validation step for domain name write operations.
	var err error
	instance.Domain, err = util.PunifySafely(instance.Domain)
	if err != nil {
		return gtserror.Newf("error punifying domain %s: %w", instance.Domain, err)
	}

	// Store the new instance model in database, invalidating cache.
	return i.state.Caches.DB.Instance.Store(instance, func() error {
		_, err := i.db.NewInsert().Model(instance).Exec(ctx)
		return err
	})
}

func (i *instanceDB) AddInstanceDeliveryError(
	ctx context.Context,
	domain string,
	errMsg string,
) error {
	// Lock to avoid overlapping
	// attempts to set/clear errors.
	i.Lock()
	defer i.Unlock()

	// Fetch instance with the given domain.
	instance, err := i.GetInstance(ctx, domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("db error getting instance: %w", err)
	}

	if instance == nil {
		// No entry so
		// nothing to do.
		return nil
	}

	// Increment error count, avoiding overflow.
	if instance.DeliveryErrorsCount != math.MaxInt16 {
		instance.DeliveryErrorsCount++
	}

	// Prepare the instance delivery error
	// to append to the instance entry.
	ide := gtsmodel.InstanceDeliveryError{
		Error: errMsg,
		Time:  time.Now(),
	}

	errLength := len(instance.DeliveryErrors)
	if errLength == 0 {
		// Create new slice.
		instance.DeliveryErrors = []gtsmodel.InstanceDeliveryError{ide}
	} else {
		// Prepend this error entry.
		instance.DeliveryErrors = slices.Insert(instance.DeliveryErrors, 0, ide)
	}

	// Trim stored errors to sensible
	// amount to avoid storing loads
	// of the same error in the db.
	const maxErrsLength = 20
	if errLength > maxErrsLength {
		instance.DeliveryErrors = instance.DeliveryErrors[:maxErrsLength]
	}

	// Update the instance entry.
	return i.state.Caches.DB.Instance.Store(instance, func() error {
		_, err := i.db.
			NewUpdate().
			Model(instance).
			Where("? = ?", bun.Ident("instance.id"), instance.ID).
			Column(
				"delivery_errors_count",
				"delivery_errors",
			).
			Exec(ctx)
		return err
	})
}

func (i *instanceDB) SetInstanceSuccessfulDelivery(
	ctx context.Context,
	domain string,
) error {
	// Lock to avoid overlapping
	// attempts to set/clear errors.
	i.Lock()
	defer i.Unlock()

	// Fetch instance with the given domain.
	instance, err := i.GetInstance(ctx, domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("db error getting instance: %w", err)
	}

	if instance == nil {
		// No entry so
		// nothing to do.
		return nil
	}

	// Set successful delivery.
	instance.LatestSuccessfulDelivery = time.Now()
	instance.DeliveryErrorsCount = 0
	instance.DeliveryErrors = nil

	// Update the instance entry.
	return i.state.Caches.DB.Instance.Store(instance, func() error {
		_, err := i.db.
			NewUpdate().
			Model(instance).
			Where("? = ?", bun.Ident("instance.id"), instance.ID).
			Column(
				"latest_successful_delivery",
				"delivery_errors_count",
				"delivery_errors",
			).
			Exec(ctx)
		return err
	})
}

func (i *instanceDB) GetInstanceAccounts(ctx context.Context, domain string, maxID string, limit int) ([]*gtsmodel.Account, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	var err error

	// Normalize the domain as punycode
	domain, err = util.Punify(domain)
	if err != nil {
		return nil, gtserror.Newf("error punifying domain %s: %w", domain, err)
	}

	// Make educated guess for slice size
	accountIDs := make([]string, 0, limit)

	q := i.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		// Select just the account ID.
		Column("account.id").
		// Select accounts belonging to given domain.
		Where("? = ?", bun.Ident("account.domain"), domain).
		Order("account.id DESC")

	if maxID == "" {
		maxID = id.Highest
	}
	q = q.Where("? < ?", bun.Ident("account.id"), maxID)

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &accountIDs); err != nil {
		return nil, err
	}

	// Catch case of no accounts early.
	count := len(accountIDs)
	if count == 0 {
		return nil, db.ErrNoEntries
	}

	// Select each account by its ID.
	accounts := make([]*gtsmodel.Account, 0, count)
	for _, id := range accountIDs {
		account, err := i.state.DB.GetAccountByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting account %q: %v", id, err)
			continue
		}

		// Append to return slice.
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (i *instanceDB) GetInstanceModeratorAddresses(ctx context.Context) ([]string, error) {
	addresses := []string{}

	// Select email addresses of approved, confirmed,
	// and enabled moderators or admins.

	q := i.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("users"), bun.Ident("user")).
		Column("user.email").
		Where("? = ?", bun.Ident("user.approved"), true).
		Where("? IS NOT NULL", bun.Ident("user.confirmed_at")).
		Where("? = ?", bun.Ident("user.disabled"), false).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("? = ?", bun.Ident("user.moderator"), true).
				WhereOr("? = ?", bun.Ident("user.admin"), true)
		}).
		OrderExpr("? ASC", bun.Ident("user.email"))

	if err := q.Scan(ctx, &addresses); err != nil {
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, db.ErrNoEntries
	}

	return addresses, nil
}

func (i *instanceDB) GetInstanceModerators(ctx context.Context) ([]*gtsmodel.Account, error) {
	accountIDs := []string{}

	// Select account IDs of approved, confirmed,
	// and enabled moderators or admins.

	q := i.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("users"), bun.Ident("user")).
		Column("user.account_id").
		Where("? = ?", bun.Ident("user.approved"), true).
		Where("? IS NOT NULL", bun.Ident("user.confirmed_at")).
		Where("? = ?", bun.Ident("user.disabled"), false).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("? = ?", bun.Ident("user.moderator"), true).
				WhereOr("? = ?", bun.Ident("user.admin"), true)
		})

	if err := q.Scan(ctx, &accountIDs); err != nil {
		return nil, err
	}

	if len(accountIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	return i.state.DB.GetAccountsByIDs(ctx, accountIDs)
}
