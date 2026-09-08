package model

import "gorm.io/gorm"

// migrateChannelProfitSchema removes only the obsolete cost-share columns.
// Historical revenue, cost and profit stay intact; upstream_ratio is never
// inferred from the old value because the pricing bases are different.
func migrateChannelProfitSchema(db *gorm.DB) error {
	for _, target := range []interface{}{&Channel{}, &ChannelProfitRecord{}} {
		columns, err := db.Migrator().ColumnTypes(target)
		if err != nil {
			return err
		}
		for _, column := range columns {
			if column.Name() == "cost_ratio" {
				if err := db.Migrator().DropColumn(target, "cost_ratio"); err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}
