package logs

// db.AddHook(&repoHook{showSql: true})

// type repoHook struct {
// 	showSql bool
// }

// func (rh *repoHook) BeforeProcess(ctx *contexts.ContextHook) (context.Context, error) {
// 	return ctx.Ctx, nil
// }

// func (rh *repoHook) AfterProcess(ctx *contexts.ContextHook) error {
// 	if ctx.Err != nil {
// 		logs.Ctx(ctx.Ctx).Caller(false).Err(ctx.Err).Str("SQL", ctx.SQL).Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Error()
// 	} else if ctx.ExecuteTime > 200*time.Millisecond {
// 		logs.Ctx(ctx.Ctx).Caller(false).Str("SlowSQL", ctx.SQL).Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Warn()
// 	} else if rh.showSql {
// 		logs.Ctx(ctx.Ctx).Caller(false).Str("SQL", ctx.SQL).Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Debug()
// 	}
// 	return ctx.Err
// }
